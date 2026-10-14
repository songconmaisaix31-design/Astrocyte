package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestRelayStoreForeignDBAndHardlinkRejectedBeforeAnyWrite(t *testing.T) {
	for _, hardlink := range []bool{false, true} {
		name := "misplaced"
		if hardlink {
			name = "hardlink"
		}
		t.Run(name, func(t *testing.T) {
			t.Setenv("LOCALAPPDATA", t.TempDir())
			root, e := RelayStateRoot()
			if e != nil {
				t.Fatal(e)
			}
			if e = os.MkdirAll(root, 0700); e != nil {
				t.Fatal(e)
			}
			foreign := filepath.Join(t.TempDir(), "simulated-cognition.sqlite")
			conn, e := sql.Open("sqlite", foreign)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = conn.Exec("CREATE TABLE attention_engine_events (id TEXT PRIMARY KEY, data TEXT); INSERT INTO attention_engine_events VALUES('public-sentinel','no raw input');"); e != nil {
				t.Fatal(e)
			}
			if e = conn.Close(); e != nil {
				t.Fatal(e)
			}
			before, e := os.ReadFile(foreign)
			if e != nil {
				t.Fatal(e)
			}
			info, e := os.Stat(foreign)
			if e != nil {
				t.Fatal(e)
			}
			target := filepath.Join(root, "relay.sqlite")
			if hardlink {
				if e = os.Link(foreign, target); e != nil {
					t.Fatal(e)
				}
			} else {
				if e = os.WriteFile(target, before, 0600); e != nil {
					t.Fatal(e)
				}
			}
			if _, e = OpenRelayStore(root); e == nil {
				t.Fatal("foreign cognition database accepted")
			}
			after, e := os.ReadFile(foreign)
			if e != nil || !bytes.Equal(before, after) {
				t.Fatal("foreign source bytes changed")
			}
			targetBytes, e := os.ReadFile(target)
			if e != nil || !bytes.Equal(before, targetBytes) {
				t.Fatal("foreign relay pathname bytes changed before rejection")
			}
			current, e := os.Stat(foreign)
			if e != nil || !current.ModTime().Equal(info.ModTime()) {
				t.Fatal("foreign source mtime changed")
			}
			for _, path := range []string{foreign, target} {
				for _, suffix := range []string{"-wal", "-shm", "-journal"} {
					if _, e = os.Stat(path + suffix); !os.IsNotExist(e) {
						t.Fatal("foreign database sidecar created before rejection")
					}
				}
			}
		})
	}
}

func TestRelayStorePathIsolationAndImmutableFirstRED(t *testing.T) {
	t.Setenv("LOCALAPPDATA", t.TempDir())
	root, e := RelayStateRoot()
	if e != nil {
		t.Fatal(e)
	}
	if _, e = OpenRelayStore(t.TempDir()); e == nil {
		t.Fatal("arbitrary source/snapshot path accepted")
	}
	s, e := OpenRelayStore(root)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	ctx := context.Background()
	red := domain.RelayReceipt{ID: "first", Status: "rejected", Input: []byte("exact private input"), Error: &domain.RelayReceiptError{Code: "scope_denied"}}
	if e = s.Update(ctx, func(st *app.RelayState) error {
		st.Receipts[red.ID] = red
		st.Workspaces["neutral"] = domain.RelayWorkspace{ID: "neutral", SourceRoot: "source-private", GitObservation: "private git observation"}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	if e = s.Update(ctx, func(st *app.RelayState) error {
		receipt := st.Receipts[red.ID]
		receipt.Status = "accepted"
		st.Receipts[red.ID] = receipt
		return nil
	}); e == nil {
		t.Fatal("first RED overwritten")
	}
	if e = s.Update(ctx, func(st *app.RelayState) error { delete(st.Receipts, red.ID); return nil }); e == nil {
		t.Fatal("first RED removed")
	}
	st, e := s.Snapshot(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if st.Receipts[red.ID].Status != "rejected" || string(st.Receipts[red.ID].Input) != "exact private input" || st.Workspaces["neutral"].SourceRoot != "source-private" || st.Workspaces["neutral"].GitObservation != "private git observation" {
		t.Fatal("dedicated private wire did not survive storage")
	}
}
