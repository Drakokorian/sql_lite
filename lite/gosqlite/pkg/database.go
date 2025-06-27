package pkg

import (
	"fmt"
	"os"
	"path/filepath"
)

// Database represents an open database connection.
// It holds references to the VFS, Pager, and other top-level components.
type Database struct {
	vfs      VFS
	pager    *Pager
	pageSize uint16
}

// Open creates a new database connection to the file at the given path.
func Open(dsn string) (*Database, error) {
	config, err := ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// For now, we only support the "os" VFS. In later phases, we will select VFS based on DSN.
	vfs := GetVFS("os")
	if vfs == nil {
		return nil, fmt.Errorf("OS VFS not registered")
	}

	const defaultCacheSize = 1024 // Number of pages in cache

	// Open the database file using the provided VFS.
	// Flags for read/write, create if not exists.
	absPath, err := filepath.Abs(config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for database file: %w", err)
	}
	file, err := vfs.Open(absPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open database file: %w", err)
	}

	fileSize, err := file.Size()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to get file size: %w", err)
	}

	var pageSize uint32
	var header *DatabaseHeader

	if fileSize == 0 { // New database file
		// Use page size from DSN if specified, otherwise default
		if config.PageSize != 0 {
			pageSize = config.PageSize
		} else {
			pageSize = 4096 // Default page size for new databases
		}
		header = DefaultDatabaseHeader(pageSize)
		// Create a temporary pager to write the header
		tempPager, err := NewPager(vfs, file, uint16(pageSize), defaultCacheSize)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to create temporary pager for new database: %w", err)
		}
		// Write the header to the first page
		headerPage := make(Page, pageSize)
		copy(headerPage, header.Bytes())
		if err := tempPager.WritePage(1, headerPage); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to write header to new database: %w", err)
		}
		if err := tempPager.FlushDirtyPages(); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to flush header to new database: %w", err)
		}
	} else {
		// Existing database, read header
		// We need a temporary pager to read the first page to get the page size
		// Assume a default page size for reading the header initially
		tempPager, err := NewPager(vfs, file, 4096, 1) // Small cache for header read
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to create temporary pager to read header: %w", err)
		}
		headerPage, err := tempPager.GetPage(1)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to read database header page: %w", err)
		}
		var actualPageSize uint32
		header, actualPageSize, err = ReadDatabaseHeader(headerPage)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to parse database header: %w", err)
		}
		pageSize = actualPageSize
	}

	// Create the actual pager with the correct page size
	pager, err := NewPager(vfs, file, uint16(pageSize), defaultCacheSize)
	if err != nil {
		file.Close() // Clean up file handle on error
		return nil, fmt.Errorf("failed to create pager: %w", err)
	}

	db := &Database{
		vfs:      vfs,
		pager:    pager,
		pageSize: uint16(pageSize),
	}

	return db, nil
}

// Close closes the database connection, flushing any pending changes to disk.
func (db *Database) Close() error {
	if db.pager == nil {
		return nil // Already closed
	}

	err := db.pager.Close()
	db.pager = nil // Mark as closed
	if err != nil {
		return fmt.Errorf("failed to close pager: %w", err)
	}

	return nil
}

// PageSize returns the page size of the database.
func (db *Database) PageSize() uint16 {
	return db.pageSize
}

// Pager returns the pager associated with the database.
func (db *Database) Pager() *Pager {
	return db.pager
}


// Database represents an open database connection.
// It holds references to the VFS, Pager, and other top-level components.
type Database struct {
	vfs      VFS
	pager    *Pager
	pageSize uint16
}

// Open creates a new database connection to the file at the given path.
func Open(dsn string) (*Database, error) {
	config, err := ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// For now, we only support the "os" VFS. In later phases, we will select VFS based on DSN.
	vfs := GetVFS("os")
	if vfs == nil {
		return nil, fmt.Errorf("OS VFS not registered")
	}

	const defaultCacheSize = 1024 // Number of pages in cache

	// Open the database file using the provided VFS.
	// Flags for read/write, create if not exists.
	absPath, err := filepath.Abs(config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for database file: %w", err)
	}
	file, err := vfs.Open(absPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open database file: %w", err)
	}

	fileSize, err := file.Size()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to get file size: %w", err)
	}

	var pageSize uint32
	var header *DatabaseHeader

	if fileSize == 0 { // New database file
		// Use page size from DSN if specified, otherwise default
		if config.PageSize != 0 {
			pageSize = config.PageSize
		} else {
			pageSize = 4096 // Default page size for new databases
		}
		header = DefaultDatabaseHeader(pageSize)
		// Create a temporary pager to write the header
		tempPager, err := NewPager(vfs, file, uint16(pageSize), defaultCacheSize)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to create temporary pager for new database: %w", err)
		}
		// Write the header to the first page
		headerPage := make(Page, pageSize)
		copy(headerPage, header.Bytes())
		if err := tempPager.WritePage(1, headerPage); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to write header to new database: %w", err)
		}
		if err := tempPager.FlushDirtyPages(); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to flush header to new database: %w", err)
		}
	} else {
		// Existing database, read header
		// We need a temporary pager to read the first page to get the page size
		// Assume a default page size for reading the header initially
		tempPager, err := NewPager(vfs, file, 4096, 1) // Small cache for header read
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to create temporary pager to read header: %w", err)
		}
		headerPage, err := tempPager.GetPage(1)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to read database header page: %w", err)
		}
		var actualPageSize uint32
		header, actualPageSize, err = ReadDatabaseHeader(headerPage)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to parse database header: %w", err)
		}
		pageSize = actualPageSize
	}

	// Create the actual pager with the correct page size
	pager, err := NewPager(vfs, file, uint16(pageSize), defaultCacheSize)
	if err != nil {
		file.Close() // Clean up file handle on error
		return nil, fmt.Errorf("failed to create pager: %w", err)
	}

	db := &Database{
		vfs:      vfs,
		pager:    pager,
		pageSize: uint16(pageSize),
	}

	return db, nil
}

// Close closes the database connection, flushing any pending changes to disk.
func (db *Database) Close() error {
	if db.pager == nil {
		return nil // Already closed
	}

	err := db.pager.Close()
	db.pager = nil // Mark as closed
	if err != nil {
		return fmt.Errorf("failed to close pager: %w", err)
	}

	return nil
}

// PageSize returns the page size of the database.
func (db *Database) PageSize() uint16 {
	return db.pageSize
}

// Pager returns the pager associated with the database.
func (db *Database) Pager() *Pager {
	return db.pager
}



