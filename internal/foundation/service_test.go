package foundation

import (
	"context"
	"testing"
)

func TestService_Health(t *testing.T) {
	svc := NewService(1)
	resp, err := svc.Health(context.Background())
	if err != nil {
		t.Fatalf("Health: %v", err)
	}

	if resp.SchemaVersion != 1 {
		t.Errorf("SchemaVersion: got %d, want 1", resp.SchemaVersion)
	}
	if resp.Status != "ok" {
		t.Errorf("Status: got %q, want ok", resp.Status)
	}
	if resp.Service != "astrocyte" {
		t.Errorf("Service: got %q, want astrocyte", resp.Service)
	}
}

func TestService_Foundation(t *testing.T) {
	svc := NewService(1)
	resp, err := svc.Foundation(context.Background())
	if err != nil {
		t.Fatalf("Foundation: %v", err)
	}

	if resp.SchemaVersion != 1 {
		t.Errorf("SchemaVersion: got %d, want 1", resp.SchemaVersion)
	}
	if resp.Stage != "S0" {
		t.Errorf("Stage: got %q, want S0", resp.Stage)
	}
	if resp.Fixture {
		t.Error("Fixture: expected false")
	}
	if resp.Storage.Engine != "sqlite" {
		t.Errorf("Storage.Engine: got %q, want sqlite", resp.Storage.Engine)
	}
	if resp.Storage.SchemaVersion != 1 {
		t.Errorf("Storage.SchemaVersion: got %d, want 1", resp.Storage.SchemaVersion)
	}

	caps := resp.Capabilities
	if caps.Imports || caps.Approvals || caps.Execution || caps.NativeResume || caps.Handoff {
		t.Error("expected all capabilities to be false in S0")
	}
}
