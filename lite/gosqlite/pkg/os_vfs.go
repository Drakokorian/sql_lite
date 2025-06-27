package pkg

import (
	"os"
	"path/filepath"
	"time"
)

// OSVFS implements the VFS interface using standard os package functions.
type OSVFS struct{}

func NewOSVFS() *OSVFS { return &OSVFS{} }

func (v *OSVFS) Open(path string, flags int, perm os.FileMode) (File, error) {
	f, err := os.OpenFile(path, flags, perm)
	if err != nil { return nil, err }
	return &OSFile{File: f}, nil
}

func (v *OSVFS) Delete(path string) error {
	return os.Remove(path)
}

func (v *OSVFS) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return false, err
}

func (v *OSVFS) Lock(path string, lockType int) error {
	// Placeholder for VFS-level locking. Actual file locking is handled by OSFile.
	return nil
}

func (v *OSVFS) Unlock(path string) error {
	// Placeholder for VFS-level unlocking. Actual file unlocking is handled by OSFile.
	return nil
}

func (v *OSVFS) CurrentTime() time.Time {
	return time.Now().UTC()
}

func (v *OSVFS) FullPath(path string) (string, error) {
	return filepath.Abs(path)
}

// OSFile wraps os.File to implement the File interface.
type OSFile struct {
	*os.File
}

func (f *OSFile) Sync() error {
	return f.File.Sync()
}

func (f *OSFile) Truncate(size int64) error {
	return f.File.Truncate(size)
}

func (f *OSFile) Size() (int64, error) {
	info, err := f.File.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// Lock implements file locking. This is a placeholder and does not provide actual locking.
// Proper platform-specific locking will be implemented in a later phase.
func (f *OSFile) Lock(lockType int) error {
	return nil
}

// Unlock implements file unlocking. This is a placeholder and does not provide actual unlocking.
// Proper platform-specific unlocking will be implemented in a later phase.
func (f *OSFile) Unlock() error {
	return nil
}

func init() {
	RegisterVFS("os", NewOSVFS())
}