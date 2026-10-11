package worksscan

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestScannerVerifiesFolderFacts(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "src")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"main.go", "README.md"} {
		if err := os.WriteFile(filepath.Join(sub, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	s := NewScanner()
	works, failures, err := s.ScanWorks(context.Background(), []string{root})
	if err != nil {
		t.Fatal(err)
	}
	if len(failures) != 0 {
		t.Fatalf("unexpected failures: %+v", failures)
	}
	if len(works) != 1 {
		t.Fatalf("expected 1 work, got %d", len(works))
	}
	work := works[0]
	if work.Kind != "folder" || !work.Files.Verified || work.Files.FileCount != 2 {
		t.Fatalf("folder facts wrong: %+v", work.Files)
	}
	if work.Files.Mtime == nil {
		t.Fatal("folder mtime must be observed")
	}
	if work.Platform != "unknown" {
		t.Fatalf("default platform must be unknown, got %q", work.Platform)
	}
}

func TestScannerReadsMarkerPlatform(t *testing.T) {
	root := t.TempDir()
	marker := `{"name":"作品A","platform":"lingxi"}`
	if err := os.WriteFile(filepath.Join(root, markerFileName), []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	s := NewScanner()
	works, failures, err := s.ScanWorks(context.Background(), []string{root})
	if err != nil {
		t.Fatal(err)
	}
	if len(works) != 1 || len(failures) != 0 {
		t.Fatalf("marker root was not observed: works=%+v failures=%+v", works, failures)
	}
	if works[0].Platform != "lingxi" || works[0].Name != "作品A" {
		t.Fatalf("marker not applied: %+v", works[0])
	}
}

func TestScannerSkipsNestedAndReportsMissing(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "node_modules", "pkg"), 0o755)
	os.WriteFile(filepath.Join(root, "node_modules", "pkg", "x.js"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(root, "a.txt"), []byte("x"), 0o644)

	s := NewScanner()
	works, failures, err := s.ScanWorks(context.Background(), []string{root, filepath.Join(root, "missing")})
	if err != nil {
		t.Fatal(err)
	}
	if len(works) != 1 || works[0].Files.FileCount != 1 {
		t.Fatalf("node_modules must be skipped, got %+v", works)
	}
	if len(failures) != 1 {
		t.Fatalf("missing root must be reported, got %+v", failures)
	}
}

func TestParseDeliveryValidatesAndAttributes(t *testing.T) {
	content := []byte(`{"works":[{"name":"app","platform":"codex","path":"C:\\w\\app","kind":"folder","attitude":"continue","handoff":"resume here"},{"name":"deck","platform":"claude","path":"C:\\w\\deck"}]}`)
	works, err := ParseDelivery(context.Background(), content)
	if err != nil {
		t.Fatal(err)
	}
	if len(works) != 2 {
		t.Fatalf("expected 2 works, got %d", len(works))
	}
	if works[0].Source != "json" || works[0].Files.Verified {
		t.Fatalf("delivery work must be unverified json source: %+v", works[0])
	}
	if works[0].Handoff.Source != "generated" {
		t.Fatalf("delivery handoff must be generated: %+v", works[0].Handoff)
	}
	if works[1].Kind != "folder" {
		t.Fatalf("empty kind must normalize to folder")
	}
}

func TestParseDeliveryRejectsInvalid(t *testing.T) {
	cases := [][]byte{
		[]byte(`{"works":[{"name":"","platform":"codex","path":"C:\\w\\app"}]}`),
		[]byte(`{"works":[{"name":"app","platform":"codex","path":""}]}`),
		[]byte(`{"works":[{"name":"app","platform":"codex","path":"C:\\w\\app","attitude":"later"}]}`),
		[]byte(`not-json`),
	}
	for i, content := range cases {
		if _, err := ParseDelivery(context.Background(), content); err == nil {
			t.Fatalf("case %d must be rejected", i)
		}
	}
}

func TestParseDeliveryHonorsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ParseDelivery(ctx, []byte(`{"works":[]}`)); err == nil {
		t.Fatal("cancelled context must error")
	}
}

func TestScannerReportsDepthBoundAsUnverified(t *testing.T) {
	root := t.TempDir()
	nested := root
	for i := 0; i <= maxDepth; i++ {
		nested = filepath.Join(nested, "nested")
	}
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "beyond.go"), []byte("private body is never read"), 0o644); err != nil {
		t.Fatal(err)
	}
	works, failures, err := NewScanner().ScanWorks(context.Background(), []string{root})
	if err != nil {
		t.Fatal(err)
	}
	if len(works) != 1 || works[0].Files.Verified || works[0].Files.Reason != "scan_depth_limit" || len(failures) != 1 {
		t.Fatalf("bounded scan must disclose unknown count: %+v %+v", works, failures)
	}
}

func TestScannerRejectsSymlinkRoot(t *testing.T) {
	root := t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(root, link); err != nil {
		t.Skipf("symlink privilege unavailable: %v", err)
	}
	works, failures, err := NewScanner().ScanWorks(context.Background(), []string{link})
	if err != nil {
		t.Fatal(err)
	}
	if len(works) != 0 || len(failures) != 1 {
		t.Fatalf("symlink must not expand approved root: %+v %+v", works, failures)
	}
}

func TestScannerRejectsSymlinkAncestor(t *testing.T) {
	target := t.TempDir()
	root := filepath.Join(target, "approved-child")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "redirect")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink privilege unavailable: %v", err)
	}
	works, failures, err := NewScanner().ScanWorks(context.Background(), []string{filepath.Join(link, "approved-child")})
	if err != nil || len(works) != 0 || len(failures) != 1 || failures[0].Reason != "root_unavailable" {
		t.Fatalf("ancestor symlink must not expand approved root: works=%+v failures=%+v error=%v", works, failures, err)
	}
}
