//go:build windows

package agents

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

func runCLIProbe(ctx context.Context, path, flag string) (string, error) {
	if err := validateProbe(path, flag); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, path, flag)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if ext := strings.ToLower(filepath.Ext(path)); ext == ".cmd" || ext == ".bat" {
		// Fixed flags and validated path only. Explicit CmdLine avoids Go's
		// CreateProcess quoting rules being interpreted as cmd.exe quoting.
		cmd = exec.CommandContext(ctx, "cmd.exe")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CmdLine: fmt.Sprintf(`cmd.exe /d /s /c ""%s" %s"`, path, flag)}
	}
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(job)
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		return "", err
	}
	output := &boundedOutput{}
	cmd.Stdout, cmd.Stderr = output, output
	cmd.WaitDelay = time.Second
	cmd.Cancel = func() error { return windows.TerminateJobObject(job, 1) }
	if err := cmd.Start(); err != nil {
		return "", err
	}
	handle, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if err == nil {
		err = windows.AssignProcessToJobObject(job, handle)
		windows.CloseHandle(handle)
	}
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return "", err
	}
	err = cmd.Wait()
	return string(output.bytes), err
}
