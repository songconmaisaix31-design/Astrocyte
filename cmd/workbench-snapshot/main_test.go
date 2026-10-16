package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	adapter "github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
)

func TestServerInitializedSnapshot(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(repo, ".verification", "snapshot-tests")
	if testRoot := os.Getenv("ASTROCYTE_TEST_OUTPUT_ROOT"); testRoot != "" {
		base = filepath.Join(testRoot, "snapshot-tests")
	}
	if err = os.MkdirAll(base, 0700); err != nil {
		t.Fatal(err)
	}
	owned, err := os.MkdirTemp(base, "actual-")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(owned, "server.exe")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/server")
	build.Dir = repo
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("server build: %s %v", output, err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	data := filepath.Join(owned, "app-state")
	server := exec.CommandContext(ctx, binary)
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "ASTROCYTE_") && !strings.HasPrefix(entry, "LOCALAPPDATA=") {
			server.Env = append(server.Env, entry)
		}
	}
	server.Env = append(server.Env, "ASTROCYTE_DATA_DIR="+data, "LOCALAPPDATA="+owned, "ASTROCYTE_PORT="+fmtInt(port), "ASTROCYTE_PROJECT_DISCOVERY_ROOTS=[]", "ASTROCYTE_ENABLE_CODEX_DISTILLATION=false", "ASTROCYTE_ENABLE_SUMMARIZE=false")
	log, err := os.Create(filepath.Join(owned, "server.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	server.Stdout = log
	server.Stderr = log
	if err = server.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Wait() }()
	defer func() {
		server.Process.Kill()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	}()
	client := http.Client{Timeout: time.Second}
	for {
		select {
		case e := <-done:
			b, _ := os.ReadFile(filepath.Join(owned, "server.log"))
			t.Fatalf("server exited: %v %s", e, b)
		default:
		}
		res, e := client.Get("http://127.0.0.1:" + fmtInt(port) + "/api/v1/health")
		if e == nil {
			res.Body.Close()
			if res.StatusCode == 200 {
				break
			}
		}
		if ctx.Err() != nil {
			b, _ := os.ReadFile(filepath.Join(owned, "server.log"))
			t.Fatalf("server readiness: %s", b)
		}
		time.Sleep(25 * time.Millisecond)
	}
	source := filepath.Join(data, "state.sqlite")
	sourceBefore, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(owned, "snapshots-root")
	p, err := publish(ctx, source, root, "")
	if err != nil {
		t.Fatal(err)
	}
	reader := adapter.NewWorkbenchReader(root)
	sourceAfter, err := os.ReadFile(source)
	if err != nil || !bytes.Equal(sourceBefore, sourceAfter) {
		t.Fatal("source main file changed during backup", err)
	}
	for _, name := range []string{"health", "today", "focus", "agents", "events", "filters", "topics/evidence"} {
		if _, err = reader.Read(ctx, name, workbench.Query{}); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	view, err := reader.Read(ctx, "events", workbench.Query{})
	if err != nil || view.Events.Total != 0 {
		t.Fatal("actual initial events must be empty", err)
	}
	before, err := os.ReadFile(filepath.Join(root, "current.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []struct{ name, source, root, hist string }{
		{"missing-source", filepath.Join(owned, "absent.sqlite"), root, ""},
		{"source-in-root", source, owned, ""},
		{"missing-history", source, root, filepath.Join(owned, "absent.json")},
	} {
		t.Run(bad.name, func(t *testing.T) {
			if _, err := publish(ctx, bad.source, bad.root, bad.hist); err == nil {
				t.Fatal("unsafe refresh succeeded")
			}
		})
	}
	t.Run("hardlink", func(t *testing.T) {
		alias := filepath.Join(root, "source-link.sqlite")
		if err := os.Link(source, alias); err != nil {
			t.Fatal(err)
		}
		if _, err := publish(ctx, source, root, ""); err == nil {
			t.Fatal("hardlink source accepted")
		}
		if err := os.Remove(alias); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("directory-alias", func(t *testing.T) {
		alias := filepath.Join(owned, "linked-root")
		if runtime.GOOS == "windows" {
			command := "New-Item -ItemType Junction -Path '" + strings.ReplaceAll(alias, "'", "''") + "' -Target '" + strings.ReplaceAll(root, "'", "''") + "' -ErrorAction Stop"
			if output, err := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-Command", command).CombinedOutput(); err != nil {
				t.Fatalf("junction creation: %v %s", err, output)
			}
		} else {
			if err := os.Symlink(root, alias); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := publish(ctx, source, alias, ""); err == nil {
			t.Fatal("directory alias accepted")
		}
	})
	t.Run("expired", func(t *testing.T) {
		expired, stop := context.WithCancel(ctx)
		stop()
		if _, err := publish(expired, source, root, ""); err == nil {
			t.Fatal("expired publish")
		}
	})
	t.Run("lock", func(t *testing.T) {
		f, err := os.OpenFile(filepath.Join(root, "snapshot.lock"), os.O_RDWR, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err = lockFile(f); err != nil {
			t.Fatal(err)
		}
		defer unlockFile(f)
		if _, err = publish(ctx, source, root, ""); err == nil {
			t.Fatal("concurrent publish accepted")
		}
	})
	after, err := os.ReadFile(filepath.Join(root, "current.json"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("failed refresh changed pointer", err)
	}
	oldPath := filepath.Join(root, "snapshots", p.Generation, "snapshot.sqlite")
	old, err := sql.Open("sqlite", uri(oldPath, true)+"&immutable=1")
	if err != nil {
		t.Fatal(err)
	}
	defer old.Close()
	var count int
	if err = old.QueryRowContext(ctx, "SELECT COUNT(*) FROM attention_engine_events").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if _, err = publish(ctx, source, root, ""); err != nil {
		t.Fatal(err)
	}
	if err = old.QueryRowContext(ctx, "SELECT COUNT(*) FROM attention_engine_events").Scan(&count); err != nil || count != 0 {
		t.Fatal("old immutable reader", err)
	}
	var h history
	if err = loadJSON(filepath.Join(root, "snapshots", p.Generation, "history.json"), &h); err != nil || len(h) != 0 {
		t.Fatal("empty history must stay empty", err)
	}
	t.Logf("actual server initialized source, production Reader 7 views, old generation %s", p.Generation)
	importRoot := filepath.Join(owned, "actual-history-import")
	if _, err = publish(ctx, source, importRoot, filepath.Join(root, "snapshots", p.Generation, "history.json")); err != nil {
		t.Fatal("explicit real prior history", err)
	}
	var out bytes.Buffer
	if err = run(ctx, []string{"--source", source, "--root", filepath.Join(owned, "cli-root"), "--timeout", "30s"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	var result struct {
		Status string `json:"status"`
		publication
	}
	if err = json.Unmarshal(out.Bytes(), &result); err != nil || result.Status != "ok" || result.Generation == "" {
		t.Fatal("CLI output", err)
	}
}

func fmtInt(n int) string { b, _ := json.Marshal(n); return string(b) }
