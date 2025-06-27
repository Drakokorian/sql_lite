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

func (s *SandboxedVFS) canonicalizeAndValidatePath(path string) (string, error) {
	// 1. Resolve symbolic links to prevent traversal exploits
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		// If symlink resolution fails, it might not be a symlink, or it's broken.
		// We proceed with the original path for now, but a more robust solution
		// might involve stricter handling or logging.
		resolvedPath = path
	}

	// 2. Get absolute path
	absPath, err := filepath.Abs(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", resolvedPath, err)
	}

	// 3. Ensure the path is clean (e.g., removes redundant slashes, resolves "./")
	cleanPath := filepath.Clean(absPath)

	// 4. Reject paths containing ".." components to prevent directory traversal
	// This check is redundant if filepath.Clean is used correctly and paths are absolute,
	// but serves as an extra layer of defense for explicit clarity.
	if strings.Contains(cleanPath, "..") && cleanPath != ".." && !strings.HasPrefix(cleanPath, "../") {
		return "", fmt.Errorf("path %s contains disallowed '..' components", path)
	}

	// 5. Disallow Windows \\?\ prefixes for security and consistency
	if runtime.GOOS == "windows" && strings.HasPrefix(cleanPath, "\\\\?\\") {
		return "", fmt.Errorf("path %s uses disallowed Windows \\\\?\\ prefix", path)
	}

	return cleanPath, nil
}

func (s *SandboxedVFS) Open(path string, flags int, perm os.FileMode) (File, error) {
	validatedPath, err := s.canonicalizeAndValidatePath(path)
	if err != nil {
		return nil, fmt.Errorf("path validation failed for %s: %w", path, err)
	}

	if _, ok := s.allowedPaths[validatedPath]; !ok {
		return nil, fmt.Errorf("access denied: %s is not an allowed path", path)
	}
	return s.baseVFS.Open(validatedPath, flags, perm)
}

func (s *SandboxedVFS) Delete(path string) error {
	validatedPath, err := s.canonicalizeAndValidatePath(path)
	if err != nil {
		return fmt.Errorf("path validation failed for %s: %w", path, err)
	}

	if _, ok := s.allowedPaths[validatedPath]; !ok {
		return fmt.Errorf("access denied: %s is not an allowed path", path)
	}
	return s.baseVFS.Delete(validatedPath)
}

func (s *SandboxedVFS) Exists(path string) (bool, error) {
	validatedPath, err := s.canonicalizeAndValidatePath(path)
	if err != nil {
		// Do not reveal existence of disallowed paths for security reasons
		return false, nil
	}

	if _, ok := s.allowedPaths[validatedPath]; !ok {
		return false, nil // Do not reveal existence of disallowed paths
	}
	return s.baseVFS.Exists(validatedPath)
}

func (s *SandboxedVFS) Lock(path string, lockType int) error {
	validatedPath, err := s.canonicalizeAndValidatePath(path)
	if err != nil {
		return fmt.Errorf("path validation failed for %s: %w", path, err)
	}

	if _, ok := s.allowedPaths[validatedPath]; !ok {
		return fmt.Errorf("access denied: %s is not an allowed path", path)
	}
	return s.baseVFS.Lock(validatedPath, lockType)
}

func (s *SandboxedVFS) Unlock(path string) error {
	validatedPath, err := s.canonicalizeAndValidatePath(path)
	if err != nil {
		return fmt.Errorf("path validation failed for %s: %w", path, err)
	}

	if _, ok := s.allowedPaths[validatedPath]; !ok {
		return fmt.Errorf("access denied: %s is not an allowed path", path)
	}
	return s.baseVFS.Unlock(validatedPath)
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
	validatedPath, err := s.canonicalizeAndValidatePath(fullPath)
	if err != nil {
		return "", fmt.Errorf("path validation failed for %s: %w", fullPath, err)
	}

	if _, ok := s.allowedPaths[validatedPath]; !ok {
		return "", fmt.Errorf("access denied: %s is not an allowed path", fullPath)
	}
	return validatedPath, nil
}
