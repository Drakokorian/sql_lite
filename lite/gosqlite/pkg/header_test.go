package pkg

import (
    "encoding/binary"
    "testing"
)

// buildValidHeader returns a byte slice of length 100 representing a minimal valid SQLite header.
func buildValidHeader(pageSize uint16) []byte {
	h := make([]byte, 100)
	copy(h[0:16], []byte("SQLite format 3\x00"))
	binary.BigEndian.PutUint16(h[16:18], pageSize)
	h[18] = 1 // write version
	h[19] = 1 // read version
	h[21] = 64
	h[22] = 32
	h[23] = 32
	// The rest can stay zero.
	return h
}

func TestReadDatabaseHeaderValid(t *testing.T) {
	buf := buildValidHeader(4096)
	dh, actual, err := ReadDatabaseHeader(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if actual != 4096 {
		t.Errorf("expected page size 4096, got %d", actual)
	}
	if string(dh.MagicString[:15]) != "SQLite format 3" {
		t.Error("magic string incorrect")
	}
}

func TestReadDatabaseHeaderInvalidMagic(t *testing.T) {
	buf := buildValidHeader(4096)
	copy(buf[0:6], []byte("BADHDR"))
	_, _, err := ReadDatabaseHeader(buf)
	if err == nil {
		t.Fatal("expected error for invalid magic string")
	}
}

func TestReadDatabaseHeaderInvalidPageSize(t *testing.T) {
	buf := buildValidHeader(123) // not power of two
	_, _, err := ReadDatabaseHeader(buf)
	if err == nil {
		t.Fatal("expected error for invalid page size")
	}
}

