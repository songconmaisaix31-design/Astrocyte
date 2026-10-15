package domain

import (
	"runtime"
	"testing"
)

func TestWriteScopesRespectLocalFilesystemCase(t *testing.T) {
	windows := runtime.GOOS == "windows"
	if got := Overlaps([]string{"src/Median.go"}, []string{"SRC/median.go"}); got != windows {
		t.Fatal("write scope case aliases do not match local filesystem semantics")
	}
	if got := ContainsPath("Src", "src/median.go"); got != windows {
		t.Fatal("parent scope case aliases do not match local filesystem semantics")
	}
	if Overlaps([]string{"src"}, []string{"src-other/median.go"}) {
		t.Fatal("distinct directory boundary was treated as an overlapping scope")
	}
}

func TestWriteScopesRejectLocalFilenameAliases(t *testing.T) {
	for _, name := range []string{"src/file.", "src/file ", "src/CON"} {
		err := ValidateWritePath(name)
		if runtime.GOOS == "windows" && err == nil {
			t.Fatalf("ambiguous Windows filename %q was accepted as a separate write scope", name)
		}
		if runtime.GOOS != "windows" && err != nil {
			t.Fatal("valid local filename was rejected", err)
		}
	}
	for _, name := range []string{".", "src", "src/median.go"} {
		if err := ValidateWritePath(name); err != nil {
			t.Fatal("normalized local scope was rejected", err)
		}
	}
}
