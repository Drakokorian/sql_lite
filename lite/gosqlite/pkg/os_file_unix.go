//go:build !windows

package pkg

import (
	"fmt"
	"os"
	"syscall"
)

// lock implements platform-specific file locking for Unix-like systems.
// It uses fcntl(F_SETLK) to acquire a lock.
func (f *OSFile) lock(lockType int) error {
	var flockType int16
	switch lockType {
	case SharedLock:
		flockType = syscall.F_RDLCK
	case ExclusiveLock:
		flockType = syscall.F_WRLCK
	default:
		return fmt.Errorf("unsupported lock type for Unix: %d", lockType)
	}

	flock := &syscall.Flock{
		Type:   flockType,
		Whence: int16(os.SEEK_SET),
		Len:    0, // Lock the entire file
	}

	// F_SETLK is non-blocking. F_SETLKW would be blocking.
	err := syscall.FcntlFlock(f.Fd(), syscall.F_SETLK, flock)
	if err != nil {
		return fmt.Errorf("failed to acquire Unix lock (type %d): %w", lockType, err)
	}
	return nil
}

// unlock implements platform-specific file unlocking for Unix-like systems.
// It uses fcntl(F_SETLK) to release a lock.
func (f *OSFile) unlock() error {
	flock := &syscall.Flock{
		Type:   syscall.F_UNLCK,
		Whence: int16(os.SEEK_SET),
		Len:    0, // Unlock the entire file
	}

	err := syscall.FcntlFlock(f.Fd(), syscall.F_SETLK, flock)
	if err != nil {
		return fmt.Errorf("failed to release Unix lock: %w", err)
	}
	return nil
}
