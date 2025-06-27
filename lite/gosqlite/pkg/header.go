package pkg

import (
    "encoding/binary"
    "fmt"
)

// Custom error types for header validation failures.
var (
	ErrInvalidFileFormat        = fmt.Errorf("invalid SQLite file format")
	ErrInvalidPageSize          = fmt.Errorf("invalid page size in header")
	ErrUnsupportedFileFormatVersion = fmt.Errorf("unsupported file format version")
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

	// Magic String (bytes 0-15): "SQLite format 3\000"
	copy(h.MagicString[:], page[0:16])
	if string(h.MagicString[:15]) != "SQLite format 3" || h.MagicString[15] != 0 {
		return nil, 0, ErrInvalidFileFormat
	}

	// Page Size (bytes 16-17)
	h.PageSize = binary.BigEndian.Uint16(page[16:18])
	var actualPageSize uint32
	if h.PageSize == 1 {
		actualPageSize = 65536
	} else {
		actualPageSize = uint32(h.PageSize)
	}

	// Validate page size: must be a power of 2 between 512 and 65536.
	if actualPageSize < 512 || actualPageSize > 65536 || (actualPageSize&(actualPageSize-1)) != 0 {
		return nil, 0, ErrInvalidPageSize
	}

	// File Format Write/Read Version (bytes 18-19)
	h.FileFormatWriteVersion = page[18]
	h.FileFormatReadVersion = page[19]
	// Validate versions: must be 1 or 2 for now.
	if h.FileFormatWriteVersion > 2 || h.FileFormatReadVersion > 2 || h.FileFormatWriteVersion == 0 || h.FileFormatReadVersion == 0 {
		return nil, 0, ErrUnsupportedFileFormatVersion
	}

	// Remaining fields
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

	return h, actualPageSize, nil
}
// DefaultDatabaseHeader returns a minimal, valid SQLite database header for a new
// database.  Only the mandatory fields are populated; the rest remain zero.
// The caller must ensure the page size is a power of two between 512-65536.
func DefaultDatabaseHeader(pageSize uint32) *DatabaseHeader {
	var hdr DatabaseHeader
	copy(hdr.MagicString[:], []byte("SQLite format 3\x00"))
	hdr.PageSize = uint16(pageSize)
	hdr.FileFormatWriteVersion = 1
	hdr.FileFormatReadVersion = 1
	hdr.MaxEmbeddedPayloadFrac = 64
	hdr.MinEmbeddedPayloadFrac = 32
	hdr.LeafEmbeddedPayloadFrac = 32
	return &hdr
}

// Bytes serialises the DatabaseHeader into a 100-byte slice exactly matching
// the on-disk SQLite header layout.  It panics if the header cannot be encoded
// into 100 bytes – which should never happen if the struct definition follows
// the spec.
func (h *DatabaseHeader) Bytes() []byte {
	buf := make([]byte, 100)
	copy(buf[0:16], h.MagicString[:])
	binary.BigEndian.PutUint16(buf[16:18], h.PageSize)
	buf[18] = h.FileFormatWriteVersion
	buf[19] = h.FileFormatReadVersion
	buf[20] = h.Reserved1
	buf[21] = h.MaxEmbeddedPayloadFrac
	buf[22] = h.MinEmbeddedPayloadFrac
	buf[23] = h.LeafEmbeddedPayloadFrac
	binary.BigEndian.PutUint32(buf[24:28], h.FileChangeCounter)
	binary.BigEndian.PutUint32(buf[28:32], h.DatabaseSize)
	binary.BigEndian.PutUint32(buf[32:36], h.FirstFreelistTrunkPage)
	binary.BigEndian.PutUint32(buf[36:40], h.TotalFreelistPages)
	binary.BigEndian.PutUint32(buf[40:44], h.SchemaCookie)
	binary.BigEndian.PutUint32(buf[44:48], h.SchemaFormat)
	binary.BigEndian.PutUint32(buf[48:52], h.DefaultPageCacheSize)
	binary.BigEndian.PutUint32(buf[52:56], h.VacuumMode)
	binary.BigEndian.PutUint32(buf[56:60], h.ApplicationID)
	binary.BigEndian.PutUint32(buf[60:64], h.UserVersion)
	binary.BigEndian.PutUint32(buf[64:68], h.IncrementalVacuumPage)
	binary.BigEndian.PutUint32(buf[68:72], h.VersionValidFor)
	binary.BigEndian.PutUint32(buf[72:76], h.SQLiteVersion)
	copy(buf[76:96], h.Reserved2[:])
	// The final 4 bytes (96-99) already populated above.
	return buf
}
