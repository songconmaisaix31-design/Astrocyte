package worksscan

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func shortPath(t *testing.T, path string) string {
	t.Helper()
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]uint16, 32768)
	n, err := windows.GetShortPathName(p, &buf[0], uint32(len(buf)))
	if err != nil || n == 0 || n >= uint32(len(buf)) {
		t.Fatalf("GetShortPathName: length=%d error=%v", n, err)
	}
	short := windows.UTF16ToString(buf[:n])
	if strings.EqualFold(short, path) {
		t.Skip("NOT_RUN: filesystem does not generate an 8.3 alias for the owned temporary root")
	}
	return short
}

func TestScannerAcceptsWindowsShortPathAliases(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "Approved long project parent")
	root := filepath.Join(parent, "Approved long project root")
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, markerFileName), []byte(`{"name":"owned alias work","platform":"codex"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "metadata.txt"), []byte("synthetic test fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	for name, alias := range map[string]string{
		"root_and_ancestors": shortPath(t, root),
		"ancestor_only":      filepath.Join(shortPath(t, parent), filepath.Base(root)),
	} {
		t.Run(name, func(t *testing.T) {
			t.Logf("actual Windows alias: %s -> %s", alias, root)
			works, failures, err := NewScanner().ScanWorks(context.Background(), []string{alias})
			if err != nil || len(failures) != 0 || len(works) != 1 {
				t.Fatalf("legal 8.3 alias rejected: works=%+v failures=%+v error=%v", works, failures, err)
			}
			if works[0].Root != alias || works[0].Platform != "codex" || works[0].Name != "owned alias work" || !works[0].Files.Verified || works[0].Files.FileCount != 2 {
				t.Fatalf("alias changed authorized root or file facts: %+v", works[0])
			}
		})
	}
}

func TestScannerRejectsWindowsJunctionAndAncestorAliases(t *testing.T) {
	target := t.TempDir()
	if err := os.Mkdir(filepath.Join(target, "child"), 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "Redirected long project junction")
	output, err := exec.Command("cmd.exe", "/c", "mklink", "/J", link, target).CombinedOutput()
	if err != nil {
		t.Fatalf("create owned junction: %v: %s", err, output)
	}
	// SameFile alone would accept this real redirect; it is not a name alias.
	linkInfo, err := os.Stat(link)
	if err != nil {
		t.Fatal(err)
	}
	targetInfo, err := os.Stat(target)
	if err != nil || !os.SameFile(linkInfo, targetInfo) {
		t.Fatalf("junction fixture does not point to target: %v", err)
	}
	for name, root := range map[string]string{
		"junction_root":     link,
		"junction_ancestor": filepath.Join(link, "child"),
	} {
		t.Run(name, func(t *testing.T) {
			works, failures, err := NewScanner().ScanWorks(context.Background(), []string{root})
			if err != nil || len(works) != 0 || len(failures) != 1 || failures[0].Reason != "root_unavailable" {
				t.Fatalf("junction must not expand approved root: works=%+v failures=%+v error=%v", works, failures, err)
			}
		})
	}
	t.Run("short_named_junction_ancestor", func(t *testing.T) {
		root := filepath.Join(shortPath(t, link), "child")
		works, failures, err := NewScanner().ScanWorks(context.Background(), []string{root})
		if err != nil || len(works) != 0 || len(failures) != 1 || failures[0].Reason != "root_unavailable" {
			t.Fatalf("short name must not hide a junction: works=%+v failures=%+v error=%v", works, failures, err)
		}
	})
}
