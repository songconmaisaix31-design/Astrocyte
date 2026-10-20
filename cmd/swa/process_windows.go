package main

import (
	"context"
	"errors"
	"fmt"
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
