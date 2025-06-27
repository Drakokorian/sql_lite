# **Phase 0: Vertical Slice (Minimum Viable Interpreter)**

**Primary Goal:** To deliver a foundational, read-only SQLite interpreter capable of executing trivial queries and demonstrating core functionality within a short timeframe (4-6 weeks).

**Success Criteria:** By the end of this phase, we will have a Go library that can:
1.  Open a `.sqlite` file.
2.  Execute `SELECT 1`.
3.  Execute `CREATE TABLE t(x)`.
4.  Execute `INSERT` statements.
5.  Print query results via a basic Command Line Interface (CLI) shell.
6.  Include comprehensive tests and a pre-built `testdata/mini.db` to validate functionality on every Pull Request (PR).

---

### **Sprint 0.1: Read-Only Pager & Baseline Interpreter**

**Objective:** To establish the absolute minimum viable execution path for SQLite operations, focusing on read-only access and a simple interpreter.

#### **Component: Read-Only Pager (`pager_readonly.go`)**

1.  **Basic Page Fetching:** Implement a simplified pager capable of reading pages from a `.sqlite` file. This pager will initially focus only on read operations, deferring write-related complexities (journaling, WAL) to later phases.
2.  **Header Integration:** Utilize the `FileHeader` parsing from Phase 1 to correctly determine page size and database file structure.
3.  **Memory Bounding (Read-Only):** Implement a basic, memory-bounded cache for read pages to prevent excessive memory consumption, even if it's a simpler eviction policy than ARC for this initial slice.

#### **Component: Baseline VDBE Interpreter (`vdbe_baseline.go`)**

1.  **Row-Wise Opcode Loop:** Implement a minimal, single-threaded, row-wise Virtual Database Engine (VDBE) interpreter. This interpreter will process opcodes one by one, operating on individual rows of data.
2.  **Build Tag for Baseline:** This interpreter will be guarded by a Go build tag (e.g., `//go:build baseline`) to allow for a clear distinction and easy selection between the baseline interpreter and future vectorized/JIT implementations.
3.  **Essential Opcodes:** Implement only the absolute minimum set of VDBE opcodes required to support:
    *   `SELECT 1` (e.g., `OP_Integer`, `OP_ResultRow`, `OP_Halt`)
    *   `CREATE TABLE t(x)` (e.g., `OP_CreateTable`, `OP_ParseSchema`)
    *   `INSERT` (e.g., `OP_Insert`, `OP_MakeRecord`)
4.  **Stubbing for Advanced Features:** For opcodes or features not yet implemented (e.g., complex joins, functions), the interpreter will return a clear, descriptive error indicating that the feature is not yet supported.

#### **Component: CLI Shell (`gosqlite_shell.go`)**

1.  **Basic Input Loop:** A simple command-line interface (CLI) that accepts SQL queries as input.
2.  **Query Execution:** Passes the input SQL query to the baseline VDBE interpreter for execution.
3.  **Result Printing:** For `SELECT` statements, prints the results to standard output in a human-readable format.
4.  **Error Reporting:** Displays any errors returned by the interpreter.

---

### **Sprint 0.2: B-Tree MVP & Metrics Integration**

**Objective:** To implement a minimal B-tree structure for table storage and integrate essential logging and metrics from day one.

#### **Component: Table B-Tree MVP (`btree_mvp.go`)**

1.  **Page Structs:** Define Go structs representing the layout of B-tree interior and leaf pages, including headers, cell pointers, and payload areas.
2.  **Cursor API:** Implement a minimal cursor API for navigating the B-tree:
    *   `First()`: Positions the cursor at the first entry in the B-tree.
    *   `Next()`: Advances the cursor to the next entry.
    *   `Seek(key)`: Positions the cursor at or after a given key.
    *   `Value()`: Retrieves the data associated with the current cursor position.
3.  **Read-Only Operations:** Focus on read-only operations for the B-tree MVP, enabling the interpreter to fetch data from existing tables.

#### **Component: Logging & Metrics (`pkg/log`, `pkg/metrics`)**

1.  **Logging Package (`pkg/log`):**
    *   **JSON Lines Format:** Implement a logging package that outputs logs in JSON Lines format for easy parsing and analysis.
    *   **Rolling File:** Configure logs to be saved securely to an external text file (e.g., `%TEMP%/gosqlite.log`) with rolling file capabilities to manage log file size.
    *   **UTC Timestamps:** Ensure all log entries include timestamps in UTC.
    *   **Critical Operations & Errors:** Log all critical operations (e.g., database open/close, transaction commit/rollback) and errors with appropriate severity levels.
2.  **Metrics Package (`pkg/metrics`):**
    *   **Integer Metrics:** Implement a metrics package capable of collecting and exposing key integer metrics.
    *   **Key Metrics:** Initially track:
        *   `pager_hit_ratio`: Cache hit ratio of the pager.
        *   `statement_latency_us`: Latency of SQL statement execution in microseconds.
        *   `wal_fsync_us`: Latency of WAL fsync operations in microseconds (will be relevant in later phases, but the metric infrastructure is in place).
    *   **Exposure:** Provide a mechanism to expose these metrics (e.g., via an in-memory counter, or a simple HTTP endpoint for debugging).

---

### **Testing Strategy for Phase 0**

*   **Unit Tests:**
    *   `pager_readonly_test.go`: Test basic page fetching and cache behavior.
    *   `vdbe_baseline_test.go`: Test execution of `SELECT 1`, `CREATE TABLE`, and `INSERT` opcodes.
    *   `btree_mvp_test.go`: Test B-tree page parsing, cursor navigation (`First`, `Next`, `Seek`), and value retrieval. Use a `testdata/mini.db` generated by `sqlite3 CLI` with 1,000 rows for scanning tests.
    *   `log_test.go`: Test JSON log formatting, file writing, and UTC timestamps.
    *   `metrics_test.go`: Test metric collection and exposure.
*   **Integration Tests:**
    *   **End-to-End CLI Test:** Write automated tests that launch the `gosqlite` shell, feed it SQL commands (including `SELECT 1`, `CREATE TABLE`, `INSERT`), and verify the output.
    *   **`testdata/mini.db` Validation:** Ship a pre-built `testdata/mini.db` file (created using the `sqlite3` CLI) alongside the tests. All integration tests will use this database to prove functionality on every PR.
*   **Fuzz Testing (Initial):**
    *   Begin basic fuzzing of the SQL input to the CLI shell and the B-tree page parsing logic to uncover unexpected panics or crashes.