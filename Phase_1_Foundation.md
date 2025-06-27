# **Phase 1: The Foundation (Hardened)**

**Primary Goal:** To build the lowest-level components for file interaction, incorporating hyper-performance I/O and a security-first VFS from day one. This phase establishes the bedrock upon which the entire hardened SQLite driver will be built.

**Success Criteria:** By the end of this phase, we will have a Go library that can:
1.  Open a `.sqlite3` file, parse its 100-byte header, and validate it as a legitimate SQLite database, adhering to strict version compatibility.
2.  Manage database configuration via a robust DSN parser.
3.  Provide a pluggable Virtual File System (VFS) interface, with a default and a sandboxed implementation.
4.  On Linux, leverage the kernel's asynchronous I/O interface for high-throughput file operations.
5.  Manage database pages in a memory-bounded cache with an advanced eviction policy.

---

### **Sprint 1.1: Database File Header, DSN Parsing & Sandboxed VFS**

**Objective:** To establish the initial connection to the database file, parse its fundamental properties, manage configuration via DSN, and introduce the secure VFS layer.

#### **Component: File Header Parser (`header.go`)**

1.  **File Format Specification:** The first 100 bytes of any SQLite 3 database file is the header. It contains critical metadata about the database's structure and properties. Our parsing must strictly adhere to the official SQLite 3 file format specification.

2.  **Go Struct Definition (`header.go`):**

    ```go
    package gosqlite

    import (
        "encoding/binary"
        "fmt"
        "io"
    )

    // FileHeader represents the 100-byte header at the beginning of an SQLite database file.
    // All multi-byte fields are stored in Big-Endian format.
    type FileHeader struct {
        MagicString         [16]byte // 0-15: Must be "SQLite format 3\000"
        PageSize            uint16   // 16-17: Database page size in bytes (e.g., 4096). Power of 2 between 512 and 32768.
        WriteVersion        byte     // 18: File format write version (1 for legacy, 2 for WAL)
        ReadVersion         byte     // 19: File format read version (1 for legacy, 2 for WAL)
        ReservedSpace       byte     // 20: Bytes of unused space at the end of each page
        MaxPayloadFrac      byte     // 21: Must be 64 (1/4 of payload)
        MinPayloadFrac      byte     // 22: Must be 32 (1/8 of payload)
        LeafPayloadFrac     byte     // 23: Must be 32 (1/8 of payload)
        FileChangeCounter   uint32   // 24-27: Incremented on each transaction
        DatabaseSizeInPages uint32   // 28-31: Size of the database file in pages
        FirstFreelistPage   uint32   // 32-35: Page number of the first freelist trunk page
        TotalFreelistPages  uint32   // 36-39: Total number of freelist pages
        SchemaCookie        uint32   // 40-43: Schema format number. Incremented on schema changes.
        SchemaFormat        uint32   // 44-47: Schema format number (1, 2, 3, or 4)
        DefaultPageCacheSize uint32  // 48-51: Default page cache size
        LargestRootBTreePage uint32 // 52-55: Page number of the largest root b-tree page
        TextEncoding        uint32   // 56-59: 1=UTF-8, 2=UTF-16le, 3=UTF-16be
        UserVersion         uint32   // 60-63: The "user version" from the PRAGMA
        IncrementalVacuum   uint32   // 64-67: True for incremental vacuum mode
        ApplicationID       uint32   // 68-71: Application ID, set by PRAGMA
        _                   [20]byte // 72-91: Reserved for expansion (must be zero)
        VersionValidFor     uint32   // 92-95: Version-valid-for number
        SQLiteVersion       uint32   // 96-99: SQLite version number (e.g., 3008007 for 3.8.7)
    }

    // ParseFileHeader reads the 100-byte header from the provided reader.
    func ParseFileHeader(r io.Reader) (*FileHeader, error) {
        header := &FileHeader{}
        buf := make([]byte, 100)
        n, err := r.Read(buf)
        if err != nil {
            return nil, fmt.Errorf("failed to read file header: %w", err)
        }
        if n != 100 {
            return nil, fmt.Errorf("expected 100 bytes for header, got %d", n)
        }

        // Manual deserialization to ensure correct byte order and validation
        copy(header.MagicString[:], buf[0:16])
        if string(header.MagicString[:15]) != "SQLite format 3" || header.MagicString[15] != 0 {
            return nil, ErrInvalidFileFormat // Custom error type
        }

        header.PageSize = binary.BigEndian.Uint16(buf[16:18])
        // Validate page size: must be a power of 2 between 512 and 32768
        if header.PageSize < 512 || header.PageSize > 32768 || (header.PageSize&(header.PageSize-1)) != 0 {
            return nil, ErrInvalidPageSize
        }

        header.WriteVersion = buf[18]
        header.ReadVersion = buf[19]
        // Validate versions (e.g., must be 1 or 2 for now)
        if header.WriteVersion > 2 || header.ReadVersion > 2 {
            return nil, ErrUnsupportedFileFormatVersion
        }

        // ... continue parsing all other fields using binary.BigEndian.Uint32 for 4-byte fields
        header.FileChangeCounter = binary.BigEndian.Uint32(buf[24:28])
        header.DatabaseSizeInPages = binary.BigEndian.Uint32(buf[28:32])
        // ... and so on for all 100 bytes

        return header, nil
    }
    ```

3.  **Error Handling:** Define specific error types for header validation failures (e.g., `ErrInvalidFileFormat`, `ErrInvalidPageSize`, `ErrUnsupportedFileFormatVersion`). These will be exported and allow for precise error handling by the calling application.

#### **Component: DSN Parser (`dsn.go`)**

1.  **DSN Format:** The driver will support a comprehensive DSN (Data Source Name) string for connection configuration, following a URL-like format. This allows for flexible and powerful configuration without modifying code.
    *   **Example:** `file:/path/to/db.sqlite?mode=rwc&cache=shared&_journal_mode=WAL&_busy_timeout=5000&_page_size=4096`

2.  **Go Struct Definition (`DSNConfig`):**

    ```go
    package gosqlite

    import (
        "net/url"
        "strconv"
        "time"
    )

    // DSNConfig holds parsed configuration parameters from the DSN string.
    type DSNConfig struct {
        Path        string
        Mode        string        // e.g., "rwc" (read/write/create), "ro" (read-only)
        Cache       string        // e.g., "shared", "private"
        JournalMode string        // e.g., "WAL", "DELETE", "TRUNCATE"
        BusyTimeout time.Duration // In milliseconds
        PageSize    uint16        // Override page size from header
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
            Path: u.Path,
            // Set sensible defaults
            Mode:        "rwc",
            Cache:       "private",
            JournalMode: "DELETE",
            BusyTimeout: 5 * time.Second,
        }

        query := u.Query()
        if m := query.Get("mode"); m != "" { config.Mode = m }
        if c := query.Get("cache"); c != "" { config.Cache = c }
        if j := query.Get("_journal_mode"); j != "" { config.JournalMode = j }
        if bt := query.Get("_busy_timeout"); bt != "" {
            ms, err := strconv.Atoi(bt)
            if err != nil { return nil, fmt.Errorf("invalid busy_timeout: %w", err) }
            config.BusyTimeout = time.Duration(ms) * time.Millisecond
        }
        if ps := query.Get("_page_size"); ps != "" {
            val, err := strconv.ParseUint(ps, 10, 16)
            if err != nil { return nil, fmt.Errorf("invalid page_size: %w", err) }
            config.PageSize = uint16(val)
        }

        return config, nil
    }
    ```

3.  **Validation:** The `ParseDSN` function will include robust validation for each parameter, returning specific errors for invalid values or unsupported options.

#### **Component: Virtual File System (VFS) (`vfs.go`)**

1.  **VFS Interface Definition:** The VFS will be a pluggable interface, allowing different implementations (standard OS, sandboxed, `io_uring`).

    ```go
    package gosqlite

    import (
        "io"
        "os"
        "time"
    )

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
    var defaultVFS VFS

    func RegisterVFS(name string, vfs VFS) {
        // Store VFS in a map, allow selection via DSN
        // For now, just set default
        defaultVFS = vfs
    }

    func GetVFS(name string) VFS {
        // Retrieve VFS by name, or return default
        return defaultVFS
    }
    ```

2.  **Standard OS VFS Implementation (`os_vfs.go`):**

    ```go
    package gosqlite

    import (
        "os"
        "path/filepath"
        "time"
    )

    // OSVFS implements the VFS interface using standard os package functions.
    type OSVFS struct{}

    func NewOSVFS() *OSVFS { return &OSVFS{} }

    func (v *OSVFS) Open(path string, flags int, perm os.FileMode) (File, error) {
        f, err := os.OpenFile(path, flags, perm)
        if err != nil { return nil, err }
        return &OSFile{File: f}, nil
    }
    // ... implement other VFS methods using os package

    // OSFile wraps os.File to implement the File interface.
    type OSFile struct {
        *os.File
    }
    // ... implement File methods, including platform-specific locking (fcntl/LockFileEx)
    ```

3.  **Sandboxed VFS Implementation (`sandboxed_vfs.go`):**

    ```go
    package gosqlite

    import (
        "fmt"
        "os"
        "path/filepath"
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
            return "", fmt.Errorf("failed to resolve symlinks for path %s: %w", path, err)
        }

        // 2. Get absolute path
        absPath, err := filepath.Abs(resolvedPath)
        if err != nil {
            return "", fmt.Errorf("failed to get absolute path for %s: %w", resolvedPath, err)
        }

        // 3. Reject paths containing ".." components to prevent directory traversal
        if strings.Contains(absPath, "..") {
            return "", fmt.Errorf("path %s contains disallowed '..' components", path)
        }

        // 4. Disallow Windows \\?\ prefixes for security and consistency
        if runtime.GOOS == "windows" && strings.HasPrefix(absPath, "\\\\?\\") {
            return "", fmt.Errorf("path %s uses disallowed Windows \\\\?\\ prefix", path)
        }

        // 5. Ensure the path is clean (e.g., removes redundant slashes)
        cleanPath := filepath.Clean(absPath)

        return cleanPath, nil
    }

    func (s *SandboxedVFS) Open(path string, flags int, perm os.FileMode) (File, error) {
        // Canonicalize and validate the requested path at runtime
        absPath, err := s.canonicalizeAndValidatePath(path)
        if err != nil {
            return nil, fmt.Errorf("path validation failed for %s: %w", path, err)
        }

        if _, ok := s.allowedPaths[absPath]; !ok {
            return nil, fmt.Errorf("access denied: %s is not an allowed path", path)
        }
        return s.baseVFS.Open(path, flags, perm)
    }
    // ... implement other VFS methods, performing path validation before delegating to baseVFS
    ```

4.  **Platform-Specific Locking:** The `OSFile` implementation will contain platform-specific code for file locking (e.g., `syscall.FcntlFlock` for Unix-like systems, `syscall.LockFileEx` for Windows). This will be managed using Go build tags (`// +build linux`, `// +build windows`).

---

### **Sprint 1.2: Asynchronous I/O (Linux-Specific)**

**Objective:** To implement a high-performance, non-blocking VFS backend for Linux utilizing the kernel's asynchronous I/O interface.

#### **Component: Asynchronous I/O VFS (`async_io_vfs.go`)**

1.  **Kernel Interface Integration:** This component will directly interact with the Linux kernel's asynchronous I/O interface (e.g., `io_uring`). It will require careful management of submission and completion queues to achieve high throughput and low latency.

    *   **Submission Queue:** A ring buffer where the application submits I/O requests (e.g., read, write, open). Each entry in this queue describes an I/O operation, including file descriptor, buffer address, length, offset, and operation type.
    *   **Completion Queue:** Another ring buffer where the kernel posts the results of completed I/O operations. The application can poll or wait on this queue to retrieve the results asynchronously.

    ```go
    // +build linux

    package gosqlite

    import (
        "fmt"
        "os"
        "syscall"
        "time"
        "unsafe"
    )

    // AsyncIOVFS implements VFS using Linux's asynchronous I/O interface.
    type AsyncIOVFS struct {
        // Internal representation of the kernel's async I/O ring
        // ... fields for managing file descriptors, pending requests, submission and completion queues
    }

    func NewAsyncIOVFS() (*AsyncIOVFS, error) {
        // Setup the kernel's async I/O ring with a specified queue depth
        // This involves system calls to initialize the ring buffers and associated kernel structures.
        // Example: ring, err := syscall.SetupAsyncIORing(256)
        // if err != nil { return nil, fmt.Errorf("async I/O setup failed: %w", err) }
        return &AsyncIOVFS{}, nil // Placeholder
    }

    func (v *AsyncIOVFS) Open(path string, flags int, perm os.FileMode) (File, error) {
        // Submit an open request to the kernel's async I/O interface
        // This would typically involve a system call equivalent to openat, but submitted asynchronously.
        // Example: fd, err := syscall.AsyncOpenat(syscall.AT_FDCWD, path, flags, uint32(perm))
        // if err != nil { return nil, err }
        return &AsyncIOFile{}, nil // Placeholder
    }

    // AsyncIOFile implements the File interface using the kernel's async I/O for reads/writes.
    type AsyncIOFile struct {
        vfs *AsyncIOVFS
        fd  int // File descriptor obtained from the kernel
        // ... fields for managing outstanding I/O requests for this file
    }

    func (f *AsyncIOFile) ReadAt(p []byte, off int64) (n int, err error) {
        // Submit a read request to the kernel's async I/O submission queue.
        // This involves populating a submission queue entry (SQE) with details
        // like file descriptor, buffer address, length, and offset.
        // The operation is non-blocking; the result will be posted to the completion queue.
        // For a synchronous API, we would then wait for the corresponding completion.
        // Example:
        // sqe := f.vfs.GetSubmissionQueueEntry()
        // sqe.Opcode = READ_OPCODE
        // sqe.Fd = int32(f.fd)
        // sqe.Addr = uint64(uintptr(unsafe.Pointer(&p[0])))
        // sqe.Len = uint32(len(p))
        // sqe.Off = uint64(off)
        // sqe.Flags |= FIXED_FILE_FLAG // Use fixed file descriptor if registered

        // Submit the request to the kernel.
        // Example: f.vfs.SubmitRequests(1)

        // Wait for completion (for synchronous API, or manage async for internal use)
        // cqe := f.vfs.WaitForCompletion()
        // if cqe.Result < 0 { return 0, fmt.Errorf("async I/O read error: %s", syscall.Errno(-cqe.Result)) }
        // return int(cqe.Result), nil
        return 0, fmt.Errorf("not implemented") // Placeholder
    }
    // ... implement WriteAt, Sync, Close, etc. using kernel async I/O opcodes
    ```

2.  **Error Handling:** Robust error handling for kernel asynchronous I/O specific errors, translating them into standard Go errors.

---

### **Sprint 1.3: Pager with Memory Bounding & ARC Cache**

**Objective:** To create a highly efficient, memory-bounded page cache that manages disk I/O and ensures data integrity.

#### **Component: Pager (`pager.go`)**

1.  **Pager Struct Definition:**

    ```go
    package gosqlite

    import (
        "fmt"
        "io"
        "sync"
    )

    type PageID uint32 // Page numbers are 1-indexed
    type Page []byte

    // Pager manages reading/writing pages from the database file and caching them.
    type Pager struct {
        vfs        VFS
        file       File
        pageSize   uint16
        dbSize     uint32 // Current size of the database in pages
        cache      *ARCCache // ARC cache for pages
        dirtyPages map[PageID]struct{} // Set of page IDs that are dirty
        mu         sync.Mutex // Mutex to protect concurrent access to pager state
        // ... other fields for journal/WAL management (in later phases)
    }

    // NewPager initializes a new Pager.
    func NewPager(vfs VFS, file File, pageSize uint16, cacheSize int) (*Pager, error) {
        p := &Pager{
            vfs:        vfs,
            file:       file,
            pageSize:   pageSize,
            cache:      NewARCCache(cacheSize), // Initialize ARC cache
            dirtyPages: make(map[PageID]struct{}),
        }
        // Read initial dbSize from file header or determine from file size
        return p, nil
    }
    ```

2.  **`GetPage(id PageID) (Page, error)`:**
    *   **Cache Lookup:** First, check the `ARCCache`. If the page is found, return it immediately.
    *   **Disk Read:** If not in cache, calculate the file offset (`(id-1) * pageSize`).
    *   **Zero-Allocation Read:** Use a pre-allocated buffer pool or `make([]byte, p.pageSize)` to minimize allocations during reads. Read the page from `p.file.ReadAt()`.
    *   **Cache Insertion:** Add the newly read page to the `ARCCache`.
    *   **Error Handling:** Handle `io.EOF` for pages beyond file size, and other I/O errors.

3.  **`WritePage(id PageID, data Page)`:**
    *   **Cache Update:** Update the page in the `ARCCache`.
    *   **Mark Dirty:** Add `id` to the `dirtyPages` set. Actual disk writes are deferred to transaction commit.
    *   **Memory Bounding:** The `ARCCache` will automatically handle eviction when its configured size limit is reached. This ensures the Pager's memory footprint remains bounded.

#### **Component: ARC Cache (`arc_cache.go`)**

1.  **ARC Algorithm:** Implement the Adaptive Replacement Cache (ARC) algorithm. ARC is a self-tuning cache replacement policy that dynamically balances between LRU (Least Recently Used) and LFU (Least Frequently Used) characteristics, generally outperforming both.

    ```go
    package gosqlite

    import "sync"

    // ARCCache implements the Adaptive Replacement Cache algorithm.
    type ARCCache struct {
        maxSize int // Maximum number of pages in the cache
        // Four LRU lists for ARC: T1, B1, T2, B2
        // T1: pages seen once recently
        // B1: pages evicted from T1
        // T2: pages seen multiple times recently
        // B2: pages evicted from T2
        // ... internal data structures (maps for quick lookup, doubly linked lists for LRU behavior)
        mu sync.Mutex
    }

    func NewARCCache(maxSize int) *ARCCache {
        // Initialize ARC lists and maps
        return &ARCCache{maxSize: maxSize}
    }

    func (c *ARCCache) Get(key PageID) (Page, bool) {
        c.mu.Lock()
        defer c.mu.Unlock()
        // ARC Get logic
        return nil, false
    }

    func (c *ARCCache) Put(key PageID, value Page) {
        c.mu.Lock()
        defer c.mu.Unlock()
        // ARC Put logic, including eviction when maxSize is exceeded
    }
    ```

2.  **Concurrency:** The `ARCCache` will be thread-safe, protected by a `sync.Mutex`.

---

### **Testing Strategy for Phase 1**

*   **Unit Tests:**
    *   `header_test.go`: Test `ParseFileHeader` with valid, invalid, and corrupted headers. Test edge cases for page size and version validation.
    *   `dsn_test.go`: Test `ParseDSN` with various valid and invalid DSN strings, ensuring correct parsing and error handling for all parameters.
    *   `vfs_test.go`: Test `OSVFS` and `SandboxedVFS` implementations. For `SandboxedVFS`, test allowed and disallowed path access attempts.
    *   `pager_test.go`: Test `NewPager`, `GetPage`, `WritePage` with mock `File` implementations. Test cache hit/miss ratios and memory bounding behavior.
    *   `arc_cache_test.go`: Comprehensive tests for the ARC algorithm's correctness, hit rates, and eviction behavior under various access patterns.
*   **Integration Tests:**
    *   Create a small, known SQLite database file (e.g., using `sqlite3` CLI). Use our `gosqlite` driver to open it, read its header, and fetch specific pages, verifying the raw byte content against expectations.
    *   Test DSN configuration end-to-end, ensuring parameters like `page_size` are correctly applied.
*   **Fuzz Testing (Initial):**
    *   Begin basic fuzzing of the `ParseFileHeader` function with random byte inputs to uncover unexpected panics or crashes.