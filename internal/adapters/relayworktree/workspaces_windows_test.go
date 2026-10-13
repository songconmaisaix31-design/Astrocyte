package relayworktree

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"golang.org/x/sys/windows"
)

func TestNeutralWindowsActualShortRoot(t *testing.T) {
	source := cleanRepo(t)
	ptr, e := windows.UTF16PtrFromString(source)
	if e != nil {
		t.Fatal(e)
	}
	buffer := make([]uint16, 32768)
	n, e := windows.GetShortPathName(ptr, &buffer[0], uint32(len(buffer)))
	if e != nil || n == 0 || n >= uint32(len(buffer)) {
		t.Skip("NOT_RUN: actual Windows short path unavailable")
	}
	short := windows.UTF16ToString(buffer[:n])
	long, e := filepath.EvalSymlinks(source)
	if e != nil {
		t.Fatal(e)
	}
	if strings.EqualFold(filepath.Clean(short), filepath.Clean(long)) {
		t.Skip("NOT_RUN: fixture has no distinct Windows 8.3 alias")
	}
	if !sameExistingDirectory(source, short) {
		t.Fatal("real short alias lost exact directory identity")
	}
	if sameExistingDirectory(source, t.TempDir()) {
		t.Fatal("unrelated directory accepted as source")
	}
	service := New(t.TempDir(), Options{AllowCommittedGit: true})
	w, e := service.Prepare(context.Background(), domain.Work{Root: short, Kind: domain.WorkKindFolder}, domain.RelayWorkspace{ID: uuid.NewString()})
	if e != nil {
		t.Fatal(e)
	}
	if !w.Prepared || w.SourceRoot != short {
		t.Fatal("approved source scope was replaced")
	}
	path := filepath.Join(w.Root, "public.txt")
	if data, e := os.ReadFile(path); e != nil || len(data) == 0 {
		t.Fatal("short root fixture did not produce a real worktree")
	}
	if _, e = service.Prepare(context.Background(), domain.Work{Root: short, Kind: domain.WorkKindFolder}, w); e != nil {
		t.Fatal(e)
	}
}
