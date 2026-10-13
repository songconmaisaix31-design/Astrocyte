package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRelayConfigurationIsolationAndNoImplicitLiveWindow(t *testing.T) {
	local := filepath.Join(os.Getenv("LOCALAPPDATA"), "astrocyte-workbench", "v0.2-access", "verification", "relay-config", uuid.NewString())
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
