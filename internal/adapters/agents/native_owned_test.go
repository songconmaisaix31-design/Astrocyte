package agents

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Test-only public helper; no installed native CLI or model is involved.
func nativeProcessTestHelper() {
	if os.Getenv("ASTROCYTE_NATIVE_PROCESS_TEST_HELPER") == "child" {
		time.Sleep(600 * time.Millisecond)
		_ = os.WriteFile(os.Getenv("ASTROCYTE_NATIVE_PROCESS_TEST_MARKER"), []byte("escaped child"), 0600)
		return
	}
	self, _ := os.Executable()
	child := exec.Command(self)
	child.Env = append(os.Environ(), "ASTROCYTE_NATIVE_PROCESS_TEST_HELPER=child")
	if err := child.Start(); err != nil {
		os.Exit(71)
	}
	fmt.Println("descendant_started")
	_ = child.Wait()
}

func TestOwnedNativeStopConfirmsExitAndKillsDescendant(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(t.TempDir(), "unwanted-child-write")
	cmd := exec.Command(self)
	cmd.Env = append(os.Environ(), "ASTROCYTE_NATIVE_PROCESS_TEST_HELPER=parent", "ASTROCYTE_NATIVE_PROCESS_TEST_MARKER="+marker)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cleanup, err := startOwnedNative(cmd)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := &nativeProcess{cmd: cmd, cancel: cancel, done: make(chan struct{}), status: "idle"}
	defer func() { _ = p.stop(context.Background()) }()
	ready := make(chan string, 1)
	go func() {
		scan := bufio.NewScanner(stdout)
		if scan.Scan() {
			ready <- scan.Text()
		}
	}()
	go func() { _ = cmd.Wait(); cleanup(); close(p.done) }()
	select {
	case line := <-ready:
		if line != "descendant_started" {
			t.Fatal(line)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("test child readiness deadline exceeded")
	}
	if p.observe().StopConfirmed {
		t.Fatal("live process was marked stopped")
	}
	if err := p.stop(ctx); err != nil {
		t.Fatal(err)
	}
	if !p.observe().StopConfirmed {
		t.Fatal("owned exit was not confirmed")
	}
	// This exceeds the child's delayed write, proving descendant termination
	// rather than merely observing the parent exit.
	time.Sleep(750 * time.Millisecond)
	if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("descendant survived owned native stop", err)
	}
}
