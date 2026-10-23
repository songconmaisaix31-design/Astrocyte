package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlitebase"
)

func TestSWAActualCompositionListenerAndShutdown(t *testing.T) {
	root := t.TempDir()
	for key, value := range map[string]string{"SWA_INSTALL_ROOT": "", "SWA_DATA_DIR": root, "SWA_PORT": "0", "ASTROCYTE_PORT": "", "SWA_WEB_DIR": "", "ASTROCYTE_WEB_DIR": "", "SWA_ALLOWED_ORIGIN": ""} {
		t.Setenv(key, value)
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()
	original := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = original }()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- runProduct(ctx, slog.New(slog.NewTextHandler(io.Discard, nil))) }()
	if err := reader.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		t.Fatal(err)
	}
	var announcement struct {
		Type   string `json:"type"`
		APIURL string `json:"api_url"`
	}
	if err := json.NewDecoder(reader).Decode(&announcement); err != nil {
		t.Fatal("actual listener announcement", err)
	}
	if announcement.Type != "http_listening" || announcement.APIURL == "" {
		t.Fatal("missing actual listener address")
	}
	response, err := http.Get(announcement.APIURL + "/api/v1/health")
	if err != nil {
		t.Fatal(err)
	}
	var health map[string]any
	err = json.NewDecoder(response.Body).Decode(&health)
	response.Body.Close()
	if err != nil || response.StatusCode != 200 || health["service"] != "astrocyte" {
		t.Fatal("actual composed health", err)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("actual service failed to shut down")
	}
	address := announcement.APIURL[len("http://"):]
	if conn, err := net.DialTimeout("tcp", address, time.Second); err == nil {
		conn.Close()
		t.Fatal("owned HTTP listener survived shutdown")
	}
	db, err := sqlitebase.Open(filepath.Join(root, "workspace", "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var count int
	if err := db.Conn().QueryRow("SELECT COUNT(*) FROM local_agent_sessions").Scan(&count); err != nil || count != 0 {
		t.Fatal("metadata startup created a native session", err)
	}
}

func TestSWAPortCompatibilityRejectsConflictingConfiguration(t *testing.T) {
	t.Setenv("SWA_PORT", "12345")
	t.Setenv("ASTROCYTE_PORT", "12346")
	if _, err := compatibleEnv("SWA_PORT", "ASTROCYTE_PORT"); err == nil {
		t.Fatal("conflicting owner and direct port accepted")
	}
	t.Setenv("SWA_PORT", "")
	if value, err := compatibleEnv("SWA_PORT", "ASTROCYTE_PORT"); err != nil || value != "12346" {
		t.Fatal("installed owner compatibility port lost", err)
	}
}
