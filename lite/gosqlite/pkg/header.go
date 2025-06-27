package pkg

import (
	"encoding/binary"
	"fmt"
)

// DatabaseHeader represents the 100-byte header at the beginning of an SQLite database file.
// For details, see: https://www.sqlite.org/fileformat2.html#database_header
type DatabaseHeader struct {
	MagicString       [16]byte // The header string: "SQLite format 3\000"
	PageSize          uint16   // Database page size in bytes. Must be a power of 2 between 512 and 65536.
	FileFormatWriteVersion uint8    // File format write version (1 for legacy, 2 for WAL).
	FileFormatReadVersion  uint8    // File format read version (1 for legacy, 2 for WAL).
	Reserved1         uint8    // Reserved for future use.
	MaxEmbeddedPayloadFrac uint8    // Maximum embedded payload fraction.
	MinEmbeddedPayloadFrac uint8    // Minimum embedded payload fraction.
	LeafEmbeddedPayloadFrac uint8    // Leaf embedded payload fraction.
	FileChangeCounter uint32   // File change counter.
	DatabaseSize      uint32   // Size of the database in pages.
	FirstFreelistTrunkPage uint32   // Page number of the first freelist trunk page.
	TotalFreelistPages uint32   // Total number of freelist pages.
	SchemaCookie      uint32   // Schema cookie.
	SchemaFormat      uint32   // Schema format number.
	DefaultPageCacheSize uint32   // Default page cache size.
	VacuumMode        uint32   // Vacuum mode (0 for none, 1 for full, 2 for incremental).
	ApplicationID     uint32   // Application ID.
	UserVersion       uint32   // User version.
	IncrementalVacuumPage uint32   // Page number of the incremental vacuum root page.
	VersionValidFor   uint32   // Version valid for.
	SQLiteVersion     uint32   // SQLite version number.
	Reserved2         [20]byte // Reserved for future use.
}

// ReadDatabaseHeader reads the 100-byte database header from the provided page.
// It returns the header struct, the actual page size (e.g., 65536, not 1), and an error.
func ReadDatabaseHeader(page Page) (*DatabaseHeader, uint32, error) {
	if len(page) < 100 {
		return nil, 0, fmt.Errorf("page too small to contain database header: %d bytes", len(page))
	}

	h := &DatabaseHeader{}

	copy(h.MagicString[:], page[0:16])
	h.PageSize = binary.BigEndian.Uint16(page[16:18]) // This reads the encoded value

	// SQLite stores 65536 as 1. Adjust for this.
	var actualPageSize uint32
	if h.PageSize == 1 {
		actualPageSize = 65536
	} else {
		actualPageSize = uint32(h.PageSize)
	}

	h.FileFormatWriteVersion = page[18]
	h.FileFormatReadVersion = page[19]
	h.Reserved1 = page[20]
	h.MaxEmbeddedPayloadFrac = page[21]
	h.MinEmbeddedPayloadFrac = page[22]
	h.LeafEmbeddedPayloadFrac = page[23]
	h.FileChangeCounter = binary.BigEndian.Uint32(page[24:28])
	h.DatabaseSize = binary.BigEndian.Uint32(page[28:32])
	h.FirstFreelistTrunkPage = binary.BigEndian.Uint32(page[32:36])
	h.TotalFreelistPages = binary.BigEndian.Uint32(page[36:40])
	h.SchemaCookie = binary.BigEndian.Uint32(page[40:44])
	h.SchemaFormat = binary.BigEndian.Uint32(page[44:48])
	h.DefaultPageCacheSize = binary.BigEndian.Uint32(page[48:52])
	h.VacuumMode = binary.BigEndian.Uint32(page[52:56])
	h.ApplicationID = binary.BigEndian.Uint32(page[56:60])
	h.UserVersion = binary.BigEndian.Uint32(page[60:64])
	h.IncrementalVacuumPage = binary.BigEndian.Uint32(page[64:68])
	h.VersionValidFor = binary.BigEndian.Uint32(page[68:72])
	h.SQLiteVersion = binary.BigEndian.Uint32(page[72:76])
	copy(h.Reserved2[:], page[76:96])

	// Basic validation of page size
	// Page size must be a power of 2 between 512 and 65536.
	isValidPageSize := false
	for _, size := range []uint32{512, 1024, 2048, 4096, 8192, 16384, 32768, 65536} {
		if actualPageSize == size {
			isValidPageSize = true
			break
		}
	}

	if !isValidPageSize {
		return nil, 0, fmt.Errorf("invalid page size in header: %d", actualPageSize)
	}

	return h, actualPageSize, nil
}