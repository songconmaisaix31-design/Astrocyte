//go:build windows

package agents

import (
	"golang.org/x/sys/windows"
	"os/exec"
	"syscall"
	"unsafe"
)

func startOwnedNative(cmd *exec.Cmd) (func(), error) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, err
	}
	cleanup := func() { _ = windows.CloseHandle(job) }
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		cleanup()
		return nil, err
	}
	cmd.Cancel = func() error { return windows.TerminateJobObject(job, 1) }
	// Exec.Cmd.Cancel is only valid with CommandContext; process lifetime is
	// controller-owned separately, so Start temporarily leaves Cancel unset.
	kill := cmd.Cancel
	cmd.Cancel = nil
	if err = cmd.Start(); err != nil {
		cleanup()
		return nil, err
	}
	cmd.Cancel = kill
	h, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if err == nil {
		err = windows.AssignProcessToJobObject(job, h)
		_ = windows.CloseHandle(h)
	}
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		cleanup()
		return nil, &nativeOwnershipError{cause: err}
	}
	return cleanup, nil
}
