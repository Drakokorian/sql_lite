package main

import (
	"fmt"
	"os"

	gosqlite "gosqlite/pkg"
)

type PageID uint32 // Page numbers are 1-indexed
type Page []byte

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gosqlite <database_file>")
		os.Exit(1)
	}

	dbPath := os.Args[1]

	// Construct a DSN string. For now, we'll just use the path.
	dsnString := fmt.Sprintf("file:./%s", dbPath)

	// Open the database.
	db, err := gosqlite.Open(dsnString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Printf("Database file '%s' opened successfully.\n", dbPath)
	fmt.Printf("Page size: %d\n", db.PageSize())

	// Example: Get page 1 (the database header)
	page, err := db.Pager().GetPage(1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting page 1: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully retrieved page 1 (header).\n")

	// In a real application, you would parse the header here.
	// For now, we just print a small part of it.
	fmt.Printf("First 16 bytes of page 1: %x\n", page[:16])

