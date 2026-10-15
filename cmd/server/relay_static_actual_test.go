package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
)

func TestRelayStaticActualMissingAndInvalidPaths(t *testing.T) {
	root := t.TempDir()
	if os.WriteFile(filepath.Join(root, "index.html"), []byte("<main>Public static content check</main>"), 0600) != nil {
		t.Fatal("实际静态文件创建失败")
	}
	server := httptest.NewServer(httpapi.NewServer(httpapi.Config{WebDir: root, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}).Handler())
	defer server.Close()
	for _, check := range []struct {
		path   string
		status int
	}{{"/", 200}, {"/workspace", 200}, {"/%00", 500}} {
		response, err := http.Get(server.URL + check.path)
		if err != nil {
			t.Fatal("真实静态 HTTP 请求失败")
		}
		if response.Body.Close() != nil || response.StatusCode != check.status {
			t.Fatalf("真实静态 HTTP 状态=%d，预期=%d", response.StatusCode, check.status)
		}
	}
}
