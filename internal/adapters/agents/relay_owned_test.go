package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRelayOwnedRootsReadOnlyDirectoryValidation(t *testing.T) {
	root := t.TempDir()
	private := filepath.Join(root, "not-a-directory")
	if err := os.WriteFile(private, []byte("public fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	validator := RelayOwnedRoots{}
	if _, err := validator.CanonicalRoot(root); err != nil {
		t.Fatal(err)
	}
	if _, err := validator.CanonicalRoot(private); err == nil {
		t.Fatal("file used as neutral root")
	}
	if _, err := validator.CanonicalRoot("relative-root"); err == nil {
		t.Fatal("relative root accepted")
	}
	link := filepath.Join(root, "redirected")
	if err := os.Symlink(t.TempDir(), link); err != nil {
		t.Log("NOT_RUN: symlink fixture unsupported on host")
		return
	}
	if _, err := validator.CanonicalRoot(link); err == nil {
		t.Fatal("redirected relay root accepted")
	}
}
