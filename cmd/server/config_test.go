package main

import (
	"path/filepath"
	"testing"
)

func TestImportRootsExplicitAndNoDefault(t *testing.T) {
	t.Setenv("ASTROCYTE_IMPORT_ROOTS", "")
	roots, err := resolveImportRoots()
	if err != nil || len(roots) != 0 {
		t.Fatalf("default roots: %v %v", roots, err)
	}
	t.Setenv("ASTROCYTE_IMPORT_ROOTS", "relative")
	if _, err := resolveImportRoots(); err == nil {
		t.Fatal("relative import root accepted")
	}
	t.Setenv("ASTROCYTE_IMPORT_ROOTS", filepath.Join(t.TempDir(), "missing"))
	if _, err := resolveImportRoots(); err == nil {
		t.Fatal("missing root accepted")
	}
	t.Setenv("ASTROCYTE_IMPORT_ROOTS", t.TempDir())
	if roots, err := resolveImportRoots(); err != nil || len(roots) != 1 {
		t.Fatalf("explicit root: %v %v", roots, err)
	}
}
func TestAllowedOriginsDoNotTrustOtherHosts(t *testing.T) {
	t.Setenv("ASTROCYTE_WEB_PORT", "15555")
	t.Setenv("ASTROCYTE_ALLOWED_ORIGINS", "")
	origins, err := resolveAllowedOrigins()
	if err != nil || origins[0] != "http://127.0.0.1:15555" {
		t.Fatalf("origin config %v %v", origins, err)
	}
	for _, origin := range []string{"http://example.com", "http://127.0.0.1.evil.test", "http://127.0.0.1:8080/path", "http://user@localhost", "file://localhost", "null"} {
		t.Setenv("ASTROCYTE_ALLOWED_ORIGINS", origin)
		if _, err := resolveAllowedOrigins(); err == nil {
			t.Fatalf("accepted %s", origin)
		}
	}
}
