package snapshot

import (
	"context"
	"syscall"
	"unsafe"
)

type memoryStatusEx struct {
	cbSize                  uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

//nolint:gochecknoglobals
var kernel = syscall.NewLazyDLL("Kernel32.dll")

// GetMemoryUsage returns current host memory consumption.
func (s *Snapshot) GetMemoryUsage(ctx context.Context) error {
	memInfo := memoryStatusEx{}
	memInfo.cbSize = uint32(unsafe.Sizeof(memInfo))
	globalmemory := kernel.NewProc("GlobalMemoryStatusEx")
	mem, _, _ := globalmemory.Call(uintptr(unsafe.Pointer(&memInfo)))

	s.System.MemFree = memInfo.ullAvailPhys
	s.System.MemUsed = memInfo.ullTotalPhys - memInfo.ullAvailPhys
	s.System.MemTotal = memInfo.ullTotalPhys

	if mem == 0 {
		return s.getMemoryUsageShared(ctx)
	}

	return nil
}
