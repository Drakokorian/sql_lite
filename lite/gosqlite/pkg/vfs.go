package pkg

import (
	"io"
	"os"
	"time"
)

type PageID uint32 // Page numbers are 1-indexed
type Page []byte

// VFS represents the Virtual File System interface for SQLite operations.
// All paths provided to VFS methods must be absolute and canonical.
type VFS interface {
	// Open opens a file at the given path with specified flags and permissions.
	Open(path string, flags int, perm os.FileMode) (File, error)
	// Delete removes a file.
	Delete(path string) error
	// Exists checks if a file exists.
	Exists(path string) (bool, error)
	// Lock acquires a file lock of the specified type.
	Lock(path string, lockType int) error
	// Unlock releases a file lock.
	Unlock(path string) error
	// CurrentTime returns the current time for file timestamps.
	CurrentTime() time.Time
	// FullPath returns the canonical absolute path for a given path.
	FullPath(path string) (string, error)
}

// File represents an open file handle within the VFS.
type File interface {
	io.ReaderAt
	io.WriterAt
	io.Closer
	io.Seeker
	Sync() error
	Truncate(size int64) error
	Size() (int64, error)
	Lock(lockType int) error // File-specific lock
	Unlock() error           // File-specific unlock
}

// Global VFS registration
var ( // Use var block for multiple declarations
	vfsRegistry = make(map[string]VFS)
	defaultVFS  VFS
)

// RegisterVFS registers a VFS implementation with a given name.
func RegisterVFS(name string, vfs VFS) {
	vfsRegistry[name] = vfs
	if defaultVFS == nil {
		defaultVFS = vfs // Set the first registered VFS as default
	}
}

// GetVFS retrieves a VFS implementation by name. If name is empty, returns the default VFS.
func GetVFS(name string) VFS {
	if name == "" {
		return defaultVFS
	}
	return vfsRegistry[name]
}

// Constants for file locking
const (
	NoLock      = 0
	SharedLock  = 1
	ReservedLock = 2
	PendingLock = 3
	ExclusiveLock = 4
)