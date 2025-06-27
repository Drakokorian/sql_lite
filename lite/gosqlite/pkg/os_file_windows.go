//go:build windows

package pkg

import (
	"fmt"
	"syscall"
	"unsafe"
)

// lock implements platform-specific file locking for Windows.
// It uses LockFileEx to acquire a lock.
func (f *OSFile) lock(lockType int) error {
	var flags uint32
	var overlapped syscall.Overlapped

	switch lockType {
	case SharedLock:
		flags = syscall.LOCKFILE_FAIL_IMMEDIATELY // Non-blocking
	case ExclusiveLock:
		flags = syscall.LOCKFILE_EXCLUSIVE_LOCK | syscall.LOCKFILE_FAIL_IMMEDIATELY // Non-blocking
	default:
		return fmt.Errorf("unsupported lock type for Windows: %d", lockType)
	}

	// Lock the entire file (0xFFFFFFFF, 0xFFFFFFFF)
	err := syscall.LockFileEx(syscall.Handle(f.Fd()), flags, 0, 0xFFFFFFFF, 0xFFFFFFFF, &overlapped)
	if err != nil {
		return fmt.Errorf("failed to acquire Windows lock (type %d): %w", lockType, err)
	}
	return nil
}

// unlock implements platform-specific file unlocking for Windows.
// It uses UnlockFileEx to release a lock.
func (f *OSFile) unlock() error {
	var overlapped syscall.Overlapped

	// Unlock the entire file (0xFFFFFFFF, 0xFFFFFFFF)
	err := syscall.UnlockFileEx(syscall.Handle(f.Fd()), 0, 0xFFFFFFFF, 0xFFFFFFFF, &overlapped)
	if err != nil {
		return fmt.Errorf("failed to release Windows lock: %w", err)
	}
	return nil
}
