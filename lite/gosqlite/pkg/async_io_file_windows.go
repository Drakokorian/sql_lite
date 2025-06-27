//go:build windows

package pkg

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Lock acquires a file-specific lock for AsyncIOFile on Windows.
// It uses LockFileEx to acquire a lock.
func (f *AsyncIOFile) Lock(lockType int) error {
	var flags uint32
	var overlapped syscall.Overlapped

	switch lockType {
	case SharedLock:
		flags = syscall.LOCKFILE_FAIL_IMMEDIATELY // Non-blocking
	case ExclusiveLock:
		flags = syscall.LOCKFILE_EXCLUSIVE_LOCK | syscall.LOCKFILE_FAIL_IMMEDIATELY // Non-blocking
	default:
		return fmt.Errorf("unsupported lock type for AsyncIOFile on Windows: %d", lockType)
	}

	// Lock the entire file (0xFFFFFFFF, 0xFFFFFFFF)
	err := syscall.LockFileEx(syscall.Handle(f.file.Fd()), flags, 0, 0xFFFFFFFF, 0xFFFFFFFF, &overlapped)
	if err != nil {
		return fmt.Errorf("failed to acquire AsyncIOFile Windows lock (type %d): %w", lockType, err)
	}
	return nil
}

// Unlock releases a file-specific lock for AsyncIOFile on Windows.
// It uses UnlockFileEx to release a lock.
func (f *AsyncIOFile) Unlock() error {
	var overlapped syscall.Overlapped

	// Unlock the entire file (0xFFFFFFFF, 0xFFFFFFFF)
	err := syscall.UnlockFileEx(syscall.Handle(f.file.Fd()), 0, 0xFFFFFFFF, 0xFFFFFFFF, &overlapped)
	if err != nil {
		return fmt.Errorf("failed to release AsyncIOFile Windows lock: %w", err)
	}
	return nil
}
