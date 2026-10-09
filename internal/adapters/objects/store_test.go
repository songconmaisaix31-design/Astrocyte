package objects

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestConcurrentPublishRestartAndCorruption(t *testing.T) {
	root := t.TempDir()
	s, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	digests := make(chan string, 16)
	for range 16 {
		wg.Go(func() {
			d, err := s.Publish(context.Background(), []byte("原文"))
			if err != nil {
				t.Error(err)
				return
			}
			digests <- d
		})
	}
	wg.Wait()
	close(digests)
	var digest string
	for d := range digests {
		if digest != "" && d != digest {
			t.Fatal("different digests")
		}
		digest = d
	}
	entries, _ := os.ReadDir(root)
	if len(entries) != 1 {
		t.Fatalf("objects = %d", len(entries))
	}
	reopened, _ := New(root)
	b, err := reopened.Read(context.Background(), digest)
	if err != nil || string(b) != "原文" {
		t.Fatalf("restart read: %s %v", b, err)
	}
	if err := os.WriteFile(filepath.Join(root, digest), []byte("corrupted"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := reopened.Publish(context.Background(), []byte("原文")); err == nil {
		t.Fatal("reused corrupt object")
	}
	if _, err := reopened.Read(context.Background(), "../state.sqlite"); err == nil {
		t.Fatal("accepted traversal")
	}
}

func TestCancelledPublication(t *testing.T) {
	s, _ := New(t.TempDir())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.Publish(ctx, []byte("x")); err == nil {
		t.Fatal("published cancelled write")
	}
	files, _ := os.ReadDir(s.root)
	if len(files) != 0 {
		t.Fatal("left an object")
	}
}
