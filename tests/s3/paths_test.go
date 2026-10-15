package s3_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

func TestActualLocalWriteScopesMatchFileIdentityAndDeviceLibrary(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".verification", "s3-paths"))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(root, "actual-")
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "Median.go")
	if err = os.WriteFile(file, []byte("Local filesystem scope identity verification.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	alias, err := os.Stat(filepath.Join(dir, "median.go"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	if err == nil && os.SameFile(info, alias) && !domain.Overlaps([]string{"Median.go"}, []string{"median.go"}) {
		t.Fatal("actual file alias was accepted as a disjoint write scope")
	}
	for _, name := range []string{"Median.go.", "Median.go "} {
		alias, err = os.Stat(filepath.Join(dir, name))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			t.Fatal(err)
		}
		if err == nil && os.SameFile(info, alias) && domain.ValidateWritePath(name) == nil {
			t.Fatal("actual dot or space file alias was accepted as an independent scope")
		}
	}
	const extended = "aux.txt"
	allowed := filepath.IsLocal(extended)
	if err = domain.ValidateWritePath(extended); (err == nil) != allowed {
		t.Fatal("extended device filename disagrees with the installed path library", err)
	}
	if allowed {
		path := filepath.Join(dir, extended)
		if err = os.WriteFile(path, []byte("This filename is supported by the actual local filesystem.\n"), 0600); err != nil {
			t.Fatal(err)
		}
		info, err = os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			t.Fatal("permitted extended device filename did not create a regular file", err)
		}
	}
}
