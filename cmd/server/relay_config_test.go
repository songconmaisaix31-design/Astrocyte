package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Linux CI has no LOCALAPPDATA. Give fixture tests an absolute, test-owned
// application-data root; production still rejects a missing/relative root.
func relayTestLocalAppData(t *testing.T) string {
	t.Helper()
	base := os.Getenv("LOCALAPPDATA")
	if !filepath.IsAbs(base) {
		base = t.TempDir()
		t.Setenv("LOCALAPPDATA", base)
	}
	return base
}

func TestRelayExplicitGuardDoesNotApplyWorkbenchDefaultAndRejectsSnapshotHardlinkBeforeWrite(t *testing.T) {
	local := filepath.Join(relayTestLocalAppData(t), "astrocyte-workbench", "v0.2-access", "verification", "relay-guard", uuid.NewString())
	if os.MkdirAll(local, 0700) != nil {
		t.Fatal("test state unavailable")
	}
	t.Setenv("LOCALAPPDATA", local)
	t.Setenv("ASTROCYTE_RELAY", "1")
	t.Setenv("ASTROCYTE_WORKBENCH", "1")
	t.Setenv("ASTROCYTE_DATA_DIR", "")
	p, e := resolveRelayDataDir("unselected")
	if e != nil {
		t.Fatal("relay isolation unavailable")
	}
	guarded, e := guardWorkbenchExplicitDataDir(p)
	if e != nil || guarded != p {
		t.Fatal("relay directory overridden by old default")
	}
	old, e := guardWorkbenchDataDir("unselected")
	if e != nil || old != filepath.Join(workbenchRoot(), "app-state") {
		t.Fatal("existing Workbench default changed")
	}
	copyDir := filepath.Join(workbenchRoot(), "snapshots", "fixture")
	for _, d := range []string{p, copyDir} {
		if os.MkdirAll(d, 0700) != nil {
			t.Fatal("fixture directory unavailable")
		}
	}
	copyFile := filepath.Join(copyDir, "snapshot.sqlite")
	before := []byte("fixture snapshot bytes")
	if os.WriteFile(copyFile, before, 0600) != nil || os.Link(copyFile, filepath.Join(p, "state.sqlite")) != nil {
		t.Fatal("hardlink fixture unavailable")
	}
	if _, e = guardWorkbenchExplicitDataDir(p); e == nil {
		t.Fatal("snapshot alias accepted for migration")
	}
	after, e := os.ReadFile(copyFile)
	if e != nil || !bytes.Equal(before, after) {
		t.Fatal("rejected alias wrote source bytes")
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		if _, e = os.Stat(copyFile + suffix); !os.IsNotExist(e) {
			t.Fatal("rejected alias created SQLite sidecar")
		}
		if _, e = os.Stat(filepath.Join(p, "state.sqlite") + suffix); !os.IsNotExist(e) {
			t.Fatal("rejected alias opened writable app DB")
		}
	}
}

func TestRelayConfigurationIsolationAndNoImplicitLiveWindow(t *testing.T) {
	local := filepath.Join(relayTestLocalAppData(t), "astrocyte-workbench", "v0.2-access", "verification", "relay-config", uuid.NewString())
	if os.MkdirAll(local, 0700) != nil {
		t.Fatal("test local directory unavailable")
	}
	t.Setenv("LOCALAPPDATA", local)
	t.Setenv("ASTROCYTE_RELAY", "1")
	t.Setenv("ASTROCYTE_DATA_DIR", "")
	t.Setenv("ASTROCYTE_RELAY_LIVE_SECONDS", "")
	root, agents, window, e := relayConfiguration()
	if e != nil || window != 0 || len(agents) != 0 {
		t.Fatal("unset registry/live policy inferred authority")
	}
	p, e := resolveRelayDataDir("unselected")
	if e != nil || !relayInside(root, p) {
		t.Fatal("application state not isolated")
	}
	t.Setenv("ASTROCYTE_DATA_DIR", "selected")
	if _, e = resolveRelayDataDir(filepath.Dir(root)); e == nil {
		t.Fatal("external application state accepted")
	}
	t.Setenv("ASTROCYTE_DATA_DIR", "")
	if os.MkdirAll(root, 0700) != nil {
		t.Fatal("test registry directory unavailable")
	}
	for _, fixture := range []struct {
		body  string
		valid bool
	}{{`[{"name":"fixture-cli","launch_cmd":["fixture-cli"],"capabilities":{},"drive_mode":"terminal"}]`, true}, {`[{"name":"fixture-cli","launch_cmd":["fixture-cli"],"drive_mode":"auto"}]`, false}, {`[{"name":"fixture-cli","launch_cmd":["fixture-cli"],"drive_mode":"terminal","actor_kind":"human"}]`, false}} {
		if os.WriteFile(filepath.Join(root, "agents.json"), []byte(fixture.body), 0600) != nil {
			t.Fatal("test registry write unavailable")
		}
		_, a, _, e := relayConfiguration()
		if (e == nil) != fixture.valid {
			t.Fatal("registration validation mismatch")
		}
		if fixture.valid && len(a) != 1 {
			t.Fatal("new data entry absent")
		}
	}
	t.Setenv("ASTROCYTE_RELAY_LIVE_SECONDS", "7")
	if os.WriteFile(filepath.Join(root, "agents.json"), []byte("[]"), 0600) != nil {
		t.Fatal("test registry write unavailable")
	}
	_, _, window, e = relayConfiguration()
	if e != nil || window != 7*time.Second {
		t.Fatal("explicit policy not preserved")
	}
	for _, value := range []string{"0", "-1", "86401", "60s"} {
		t.Setenv("ASTROCYTE_RELAY_LIVE_SECONDS", value)
		if _, _, _, e = relayConfiguration(); e == nil {
			t.Fatal("invalid live window accepted")
		}
	}
	t.Setenv("ASTROCYTE_RELAY", "0")
	if p, e := resolveRelayDataDir("existing-app"); e != nil || p != "existing-app" {
		t.Fatal("disabled access changed existing configuration")
	}
}

func TestRelayProductionStillRejectsMissingOrRelativeLocalAppData(t *testing.T) {
	for _, value := range []string{"", "relative-local-app-data"} {
		t.Setenv("LOCALAPPDATA", value)
		if _, e := relayRoot(); e == nil {
			t.Fatal("production accepted implicit or relative application state")
		}
	}
}
