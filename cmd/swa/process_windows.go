package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type resourceSample struct {
	UTC            time.Time `json:"utc"`
	CPUPercent     float64   `json:"cpu_percent"`
	AvailableBytes uint64    `json:"available_bytes"`
	ElapsedSeconds float64   `json:"elapsed_seconds"`
}

func configureCurrentProcess() error {
	return windows.SetPriorityClass(windows.CurrentProcess(), windows.BELOW_NORMAL_PRIORITY_CLASS)
}

func configureServiceChild(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.BELOW_NORMAL_PRIORITY_CLASS}
}

type serviceJob struct{ handle windows.Handle }

func ownServiceTree() (*serviceJob, error) {
	handle, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, err
	}
	job := &serviceJob{handle: handle}
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	// 当前 serve 在创建子进程前加入独立 Job；关闭句柄不终止进程。
	if _, err := windows.SetInformationJobObject(handle, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		return nil, errors.Join(err, job.close())
	}
	if err := windows.AssignProcessToJobObject(handle, windows.CurrentProcess()); err != nil {
		return nil, errors.Join(err, job.close())
	}
	active, err := job.active()
	if err != nil || active != 1 {
		return nil, errors.Join(fmt.Errorf("serve Job 初始成员不确定: active=%d", active), err, job.close())
	}
	return job, nil
}

func (job *serviceJob) active() (uint32, error) {
	if err := job.member(windows.CurrentProcess()); err != nil {
		return 0, err
	}
	var accounting struct {
		UserTime, KernelTime, PeriodUserTime, PeriodKernelTime  int64
		PageFaults, TotalProcesses, ActiveProcesses, Terminated uint32
	}
	var length uint32
	if err := windows.QueryInformationJobObject(job.handle, windows.JobObjectBasicAccountingInformation, uintptr(unsafe.Pointer(&accounting)), uint32(unsafe.Sizeof(accounting)), &length); err != nil {
		return 0, err
	}
	if length != uint32(unsafe.Sizeof(accounting)) || accounting.ActiveProcesses < 1 || accounting.TotalProcesses < accounting.ActiveProcesses {
		return 0, errors.New("serve Job 实际 accounting 无效")
	}
	return accounting.ActiveProcesses, nil
}

func (job *serviceJob) member(process windows.Handle) error {
	var member int32
	call := windows.NewLazySystemDLL("kernel32.dll").NewProc("IsProcessInJob")
	result, _, err := call.Call(uintptr(process), uintptr(job.handle), uintptr(unsafe.Pointer(&member)))
	if result == 0 {
		return fmt.Errorf("IsProcessInJob: %w", err)
	}
	if member == 0 {
		return errors.New("实际进程不属于本次 serve Job")
	}
	return nil
}

func (job *serviceJob) observeChild(command *exec.Cmd, stderr io.Writer) error {
	process, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, uint32(command.Process.Pid))
	if err != nil {
		return err
	}
	observe := func() error {
		if err := job.member(process); err != nil {
			return err
		}
		var created, exited, kernel, user windows.Filetime
		if err := windows.GetProcessTimes(process, &created, &exited, &kernel, &user); err != nil {
			return err
		}
		active, err := job.active()
		if err != nil {
			return err
		}
		if created.Nanoseconds() <= 0 || active <= 1 {
			return errors.New("所属子进程的创建身份或非空 Job 成员未确认")
		}
		_, err = fmt.Fprintf(stderr, "serve Job 子进程: pid=%d created_utc=%s active=%d\n", command.Process.Pid, time.Unix(0, created.Nanoseconds()).UTC().Format(time.RFC3339Nano), active)
		return err
	}
	return errors.Join(observe(), windows.CloseHandle(process))
}

func (job *serviceJob) awaitDescendants(stderr io.Writer) error {
	deadline := time.NewTimer(15 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	consecutive := 0
	for {
		active, err := job.active()
		if err != nil {
			return err
		}
		if active == 1 {
			consecutive++
			_, _ = fmt.Fprintf(stderr, "serve Job 退出观察: utc=%s active=%d consecutive=%d\n", time.Now().UTC().Format(time.RFC3339Nano), active, consecutive)
			if consecutive == 2 {
				return nil
			}
		} else {
			consecutive = 0
		}
		select {
		case <-deadline.C:
			return fmt.Errorf("serve Job 仍有所属后代: active=%d", active)
		case <-tick.C:
		}
	}
}

func (job *serviceJob) close() error { return windows.CloseHandle(job.handle) }

type cpuTimes struct{ idle, kernel, user windows.Filetime }

func readCPUTimes() (cpuTimes, error) {
	var times cpuTimes
	call := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetSystemTimes")
	result, _, err := call.Call(uintptr(unsafe.Pointer(&times.idle)), uintptr(unsafe.Pointer(&times.kernel)), uintptr(unsafe.Pointer(&times.user)))
	if result == 0 {
		return times, fmt.Errorf("GetSystemTimes: %w", err)
	}
	return times, nil
}

func ticks(value windows.Filetime) uint64 {
	return uint64(value.HighDateTime)<<32 | uint64(value.LowDateTime)
}

func resourceAdmission(ctx context.Context) ([]resourceSample, error) {
	if runtime.NumCPU() > 64 {
		return nil, errors.New("当前 GetSystemTimes 采样仅接受不超过 64 个逻辑处理器")
	}
	started := time.Now()
	samples := make([]resourceSample, 0, 2)
	for index := 0; index < 2; index++ {
		if index > 0 {
			select {
			case <-ctx.Done():
				return samples, ctx.Err()
			case <-time.After(1500 * time.Millisecond):
			}
		}
		before, err := readCPUTimes()
		if err != nil {
			return samples, err
		}
		select {
		case <-ctx.Done():
			return samples, ctx.Err()
		case <-time.After(time.Second):
		}
		after, err := readCPUTimes()
		if err != nil {
			return samples, err
		}
		idle := ticks(after.idle) - ticks(before.idle)
		total := ticks(after.kernel) - ticks(before.kernel) + ticks(after.user) - ticks(before.user)
		if total == 0 || idle > total {
			return samples, errors.New("实际系统 CPU 计时不能形成有效阻塞样本")
		}
		var memory struct {
			Length, Load                                                       uint32
			TotalPhysical, AvailablePhysical, TotalPageFile, AvailablePageFile uint64
			TotalVirtual, AvailableVirtual, AvailableExtendedVirtual           uint64
		}
		memory.Length = uint32(unsafe.Sizeof(memory))
		call := windows.NewLazySystemDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")
		result, _, err := call.Call(uintptr(unsafe.Pointer(&memory)))
		if result == 0 {
			return samples, fmt.Errorf("GlobalMemoryStatusEx: %w", err)
		}
		samples = append(samples, resourceSample{UTC: time.Now().UTC(), CPUPercent: 100 * float64(total-idle) / float64(total), AvailableBytes: memory.AvailablePhysical, ElapsedSeconds: time.Since(started).Seconds()})
	}
	interval := samples[1].ElapsedSeconds - samples[0].ElapsedSeconds
	if interval < 2 || interval > 5 {
		return samples, errors.New("两次实际样本间隔超出 2–5 秒")
	}
	for _, sample := range samples {
		if sample.CPUPercent >= 85 || sample.AvailableBytes < 4*1024*1024*1024 {
			return samples, errors.New("实际启动样本未满足 CPU 小于 85% 及可用内存至少 4 GiB")
		}
	}
	return samples, nil
}
