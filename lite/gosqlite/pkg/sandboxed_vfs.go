package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// SandboxedVFS wraps another VFS and restricts file access to a predefined set of allowed paths.
type SandboxedVFS struct {
	baseVFS     VFS
	allowedPaths map[string]struct{}
}

func NewSandboxedVFS(base VFS, allowed ...string) *SandboxedVFS {
	s := &SandboxedVFS{baseVFS: base, allowedPaths: make(map[string]struct{})}
	for _, p := range allowed {
		// Canonicalize and validate paths during initialization
		absPath, err := s.canonicalizeAndValidatePath(p)
		if err != nil {
			// Log or handle error during initialization if a provided path is invalid
			// For now, we'll just print to stderr and skip it.
			fmt.Fprintf(os.Stderr, "Warning: Invalid allowed path provided to SandboxedVFS: %s, error: %v\n", p, err)
			continue
		}
		s.allowedPaths[absPath] = struct{}{}
	}
	return s
}

func (s *SandboxedVFS) isPathAllowed(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}

	cleanPath := filepath.Clean(absPath)

	for allowed := range s.allowedPaths {
		if strings.HasPrefix(cleanPath, allowed) {
			return cleanPath, nil
		}
	}

	return "", fmt.Errorf("path %s is not in allowed directories", path)
}

func (s *SandboxedVFS) Open(path string, flags int, perm os.FileMode) (File, error) {
	allowed, err := s.isPathAllowed(path)
	if err != nil {
		return nil, fmt.Errorf("path validation failed for %s: %w", path, err)
	}
	if !allowed {
		return nil, fmt.Errorf("access denied: %s is not an allowed path", path)
	}
	return s.baseVFS.Open(path, flags, perm)
}

func (s *SandboxedVFS) Delete(path string) error {
	allowed, err := s.isPathAllowed(path)
	if err != nil {
		return fmt.Errorf("path validation failed for %s: %w", path, err)
	}
	if !allowed {
		return fmt.Errorf("access denied: %s is not an allowed path", path)
	}
	return s.baseVFS.Delete(path)
}

func (s *SandboxedVFS) Exists(path string) (bool, error) {
	allowed, err := s.isPathAllowed(path)
	if err != nil {
		return false, fmt.Errorf("path validation failed for %s: %w", path, err)
	}
	if !allowed {
		return false, nil // Do not reveal existence of disallowed paths
	}
	return s.baseVFS.Exists(path)
}

func (s *SandboxedVFS) Lock(path string, lockType int) error {
	allowed, err := s.isPathAllowed(path)
	if err != nil {
		return fmt.Errorf("path validation failed for %s: %w", path, err)
	}
	if !allowed {
		return fmt.Errorf("access denied: %s is not an allowed path", path)
	}
	return s.baseVFS.Lock(path, lockType)
}

func (s *SandboxedVFS) Unlock(path string) error {
	allowed, err := s.isPathAllowed(path)
	if err != nil {
		return fmt.Errorf("path validation failed for %s: %w", path, err)
	}
	if !allowed {
		return fmt.Errorf("access denied: %s is not an allowed path", path)
	}
	return s.baseVFS.Unlock(path)
}

func (s *SandboxedVFS) CurrentTime() time.Time {
	return s.baseVFS.CurrentTime()
}

func (s *SandboxedVFS) FullPath(path string) (string, error) {
	// FullPath should return the canonical path, but we still need to validate it.
	fullPath, err := s.baseVFS.FullPath(path)
	if err != nil {
		return "", err
	}
	allowed, err := s.isPathAllowed(fullPath)
	if err != nil {
		return "", fmt.Errorf("path validation failed for %s: %w", fullPath, err)
	}
	if !allowed {
		return "", fmt.Errorf("access denied: %s is not an allowed path", fullPath)
	}
	return fullPath, nil
}