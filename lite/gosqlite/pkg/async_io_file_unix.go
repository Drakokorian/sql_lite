//go:build !windows

package pkg

import (
	"fmt"
	"syscall"
)

// Lock acquires a file-specific lock for AsyncIOFile on Unix-like systems.
// It uses fcntl(F_SETLK) to acquire a lock, ensuring non-blocking behavior.
func (f *AsyncIOFile) Lock(lockType int) error {
	var flockType int16
	switch lockType {
	case SharedLock:
		flockType = syscall.F_RDLCK
	case ExclusiveLock:
		flockType = syscall.F_WRLCK
	default:
		return fmt.Errorf("unsupported lock type for AsyncIOFile on Unix: %d", lockType)
	}

	flock := &syscall.Flock{
		Type:   flockType,
		Whence: int16(os.SEEK_SET),
		Len:    0, // Lock the entire file
	}

	// F_SETLK is non-blocking. This call attempts to acquire the lock immediately.
	err := syscall.FcntlFlock(f.file.Fd(), syscall.F_SETLK, flock)
	if err != nil {
		return fmt.Errorf("failed to acquire AsyncIOFile Unix lock (type %d): %w", lockType, err)
	}
	return nil
}

// Unlock releases a file-specific lock for AsyncIOFile on Unix-like systems.
// It uses fcntl(F_SETLK) to release a lock.
func (f *AsyncIOFile) Unlock() error {
	flock := &syscall.Flock{
		Type:   syscall.F_UNLCK,
		Whence: int16(os.SEEK_SET),
		Len:    0, // Unlock the entire file
	}

	err := syscall.FcntlFlock(f.file.Fd(), syscall.F_SETLK, flock)
	if err != nil {
		return fmt.Errorf("failed to release AsyncIOFile Unix lock: %w", err)
	}
	return nil
}
