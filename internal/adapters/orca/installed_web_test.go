package orca

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
)

func installedWebInput(t *testing.T) (string, string) {
	t.Helper()
	install := t.TempDir()
	web := filepath.Join(install, "web")
	for _, path := range []string{filepath.Join(web, "assets"), filepath.Join(web, ".vite")} {
		if err := os.MkdirAll(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for path, body := range map[string]string{
		filepath.Join(install, "EDITION"):            "combined\n",
		filepath.Join(web, "index.html"):             "<!doctype html><title>SWA public test</title>",
		filepath.Join(web, "assets", "index.js"):     "export {};\n",
		filepath.Join(web, ".vite", "manifest.json"): `{"index.html":{"file":"assets/index.js","isEntry":true}}`,
	} {
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return install, web
}

func TestInstalledWebDirectoryAndHTTPRoutes(t *testing.T) {
	install, web := installedWebInput(t)
	actual, err := InstalledWebDirectory(install, web)
	if err != nil || actual != web {
		t.Fatalf("trusted installed Web directory: %q %v", actual, err)
	}
	if actual, err := InstalledWebDirectory("", ""); err != nil || actual != "" {
		t.Fatal("CLI without an explicit Web directory must remain available", err)
	}
	server := httpapi.NewServer(httpapi.Config{WebDir: actual})
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()
	for _, path := range []string{"/attention", "/assets/index.js"} {
		response, err := http.Get(httpServer.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil || response.StatusCode != http.StatusOK || len(body) == 0 {
			t.Fatalf("installed asset %s: HTTP %d %v", path, response.StatusCode, err)
		}
	}
	for _, path := range []string{"/api", "/api/v1/does-not-exist"} {
		response, err := http.Get(httpServer.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil || response.StatusCode != http.StatusNotFound || !strings.Contains(response.Header.Get("Content-Type"), "application/json") || strings.Contains(string(body), "<!doctype") {
			t.Fatalf("API must retain its actual JSON error: %s HTTP %d %s %v", path, response.StatusCode, body, err)
		}
	}
}

func TestInstalledWebDirectoryRejectsUntrustedInputs(t *testing.T) {
	cases := []string{"wrong-root", "missing-edition", "cli-edition", "missing-index", "missing-manifest", "invalid-manifest", "wrong-shape", "missing-asset", "escape-file", "escape-css", "escape-assets", "missing-import", "symlink-file", "symlink-directory", "symlink-install"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			install, web := installedWebInput(t)
			write := func(path, text string) {
				t.Helper()
				if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			remove := func(path string) {
				t.Helper()
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			}
			link := func(target, path string) {
				t.Helper()
				if err := os.Symlink(target, path); err != nil {
					t.Skipf("actual platform cannot create this symbolic link: %v", err)
				}
			}
			manifest := filepath.Join(web, ".vite", "manifest.json")
			switch name {
			case "wrong-root":
				web = t.TempDir()
			case "missing-edition":
				remove(filepath.Join(install, "EDITION"))
			case "cli-edition":
				write(filepath.Join(install, "EDITION"), "cli\n")
			case "missing-index":
				remove(filepath.Join(web, "index.html"))
			case "missing-manifest":
				remove(manifest)
			case "invalid-manifest":
				write(manifest, "{")
			case "wrong-shape":
				write(manifest, `{"entry":{"file":"assets/index.js"}}`)
			case "missing-asset":
				remove(filepath.Join(web, "assets", "index.js"))
			case "escape-file":
				write(manifest, `{"index.html":{"file":"../EDITION","isEntry":true}}`)
			case "escape-css":
				write(manifest, `{"index.html":{"file":"assets/index.js","isEntry":true,"css":["../EDITION"]}}`)
			case "escape-assets":
				write(manifest, `{"index.html":{"file":"assets/index.js","isEntry":true,"assets":["C:\\private"]}}`)
			case "missing-import":
				write(manifest, `{"index.html":{"file":"assets/index.js","isEntry":true,"imports":["missing"]}}`)
			case "symlink-file":
				remove(filepath.Join(web, "assets", "index.js"))
				link(filepath.Join(install, "EDITION"), filepath.Join(web, "assets", "index.js"))
			case "symlink-directory":
				link(t.TempDir(), filepath.Join(web, "alias"))
			case "symlink-install":
				alias := filepath.Join(t.TempDir(), "alias")
				link(install, alias)
				install, web = alias, filepath.Join(alias, "web")
			}
			if _, err := InstalledWebDirectory(install, web); err == nil {
				t.Fatal("untrusted installed Web input was accepted")
			}
		})
	}
}
