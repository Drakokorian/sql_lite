// +build !linux

package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AsyncIOVFS is a stub for non-Linux systems.
type AsyncIOVFS struct{}

// NewAsyncIOVFS returns an error on non-Linux systems.
func NewAsyncIOVFS() (*AsyncIOVFS, error) {
	return nil, fmt.Errorf("AsyncIOVFS is only available on Linux")
}

func (v *AsyncIOVFS) Open(path string, flags int, perm os.FileMode) (File, error) {
	return nil, fmt.Errorf("AsyncIOVFS not available on this OS")
}

func (v *AsyncIOVFS) Delete(path string) error {
	return fmt.Errorf("AsyncIOVFS not available on this OS")
}

func (v *AsyncIOVFS) Exists(path string) (bool, error) {
	return false, fmt.Errorf("AsyncIOVFS not available on this OS")
}

func (v *AsyncIOVFS) Lock(path string, lockType int) error {
	return fmt.Errorf("AsyncIOVFS not available on this OS")
}

func (v *AsyncIOVFS) Unlock(path string) error {
	return fmt.Errorf("AsyncIOVFS not available on this OS")
}

func (v *AsyncIOVFS) CurrentTime() time.Time {
	return time.Now().UTC()
}

func (v *AsyncIOVFS) FullPath(path string) (string, error) {
	return filepath.Abs(path)
}

type AsyncIOFile struct{}

func (f *AsyncIOFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, fmt.Errorf("AsyncIOVFS not available on this OS")
}

func (f *AsyncIOFile) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, fmt.Errorf("AsyncIOVFS not available on this OS")
}

func (f *AsyncIOFile) Close() error {
	return fmt.Errorf("AsyncIOVFS not available on this OS")
}

func (f *AsyncIOFile) Sync() error {
	return fmt.Errorf("AsyncIOVFS not available on this OS")
}

func (f *AsyncIOFile) Truncate(size int64) error {
	return fmt.Errorf("AsyncIOVFS not available on this OS")
}

func (f *AsyncIOFile) Size() (int64, error) {
	return 0, fmt.Errorf("AsyncIOVFS not available on this OS")
}

func (f *AsyncIOFile) Lock(lockType int) error {
	return fmt.Errorf("AsyncIOVFS not available on this OS")
}

func (f *AsyncIOFile) Unlock() error {
	return fmt.Errorf("AsyncIOVFS not available on this OS")
}
