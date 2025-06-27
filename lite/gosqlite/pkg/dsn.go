package pkg

import (
	"fmt"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// DSNConfig holds parsed configuration parameters from the DSN string.
type DSNConfig struct {
	Path        string
	VFS         string        // Optional VFS name (e.g., "sandbox", "os")
	Mode        string        // e.g., "rwc" (read/write/create), "ro" (read-only)
	Cache       string        // e.g., "shared", "private"
	JournalMode string        // e.g., "WAL", "DELETE", "TRUNCATE"
	BusyTimeout time.Duration // In milliseconds
	PageSize    uint32        // Override page size from header
	Synchronous string        // e.g., "FULL", "NORMAL", "OFF"
	ForeignKeys bool          // Enable or disable foreign key constraints
	// ... other parameters like synchronous, foreign_keys, etc.
}

// ParseDSN parses a DSN string into a DSNConfig struct.
func ParseDSN(dsn string) (*DSNConfig, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN format: %w", err)
	}

	if u.Scheme != "file" {
		return nil, fmt.Errorf("unsupported DSN scheme: %s", u.Scheme)
	}

	config := &DSNConfig{
		// Determine the path based on URL parsing behavior
		Path: func() string {
			if u.Opaque != "" { // For DSNs like "file:test.db"
				return u.Opaque
			} else if u.Host != "" && runtime.GOOS == "windows" { // For Windows absolute paths like "file:///C:/path/to/db"
				// url.Parse on Windows might put the drive letter in Host and the rest in Path
				return u.Host + u.Path
			} else { // For DSNs like "file:///path/to/db" or "file:./test.db"
				return u.Path
			}
		}(),
		// Set sensible defaults
		Mode:        "rwc",
		Cache:       "private",
		JournalMode: "DELETE",
		BusyTimeout: 5 * time.Second,
		Synchronous: "FULL",
		ForeignKeys: false,
	}

	query := u.Query()
	if m := query.Get("mode"); m != "" {
		switch strings.ToLower(m) {
		case "ro", "rw", "rwc", "memory":
			config.Mode = m
		default:
			return nil, fmt.Errorf("invalid mode: %s", m)
		}
	}
	if c := query.Get("cache"); c != "" {
		switch strings.ToLower(c) {
		case "shared", "private":
			config.Cache = c
		default:
			return nil, fmt.Errorf("invalid cache: %s", c)
		}
	}
	if j := query.Get("_journal_mode"); j != "" {
		switch strings.ToUpper(j) {
		case "DELETE", "TRUNCATE", "PERSIST", "MEMORY", "WAL", "OFF":
			config.JournalMode = j
		default:
			return nil, fmt.Errorf("invalid _journal_mode: %s", j)
		}
	}
	if bt := query.Get("_busy_timeout"); bt != "" {
		ms, err := strconv.Atoi(bt)
		if err != nil {
			return nil, fmt.Errorf("invalid _busy_timeout: %w", err)
		}
		config.BusyTimeout = time.Duration(ms) * time.Millisecond
	}
	if ps := query.Get("_page_size"); ps != "" {
		val, err := strconv.ParseUint(ps, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid _page_size: %w", err)
		}
		// Validate page size: must be a power of 2 between 512 and 65536
		if val < 512 || val > 65536 || (val&(val-1)) != 0 {
			return nil, fmt.Errorf("page size must be a power of 2 between 512 and 65536")
		}
		config.PageSize = uint32(val)
	}
	if s := query.Get("_synchronous"); s != "" {
		s = strings.ToUpper(s)
		switch s {
		case "FULL", "NORMAL", "OFF":
			config.Synchronous = s
		default:
			return nil, fmt.Errorf("invalid _synchronous: %s", s)
		}
	}
	if fk := query.Get("_foreign_keys"); fk != "" {
		val, err := strconv.ParseBool(fk)
		if err != nil {
			return nil, fmt.Errorf("invalid _foreign_keys: %w", err)
		}
		config.ForeignKeys = val
	}

	return config, nil
}