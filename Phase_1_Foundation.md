Phase 1 — Pure-Go Storage Foundation
(No CGO, no sqlite3.c, no external packages—100 % Go std-lib + syscalls)

Area	“Fully functional” deliverables	Validation (all with go test, std-lib only)
Header integrity	• Read exactly 100 bytes at file offset 0.	
• Verify: magic string SQLite format 3\0, page-size ∈ {512 … 32768} & power-of-2, read/write versions ≤ 2, reserved bytes = 0.		
• Return typed errors: ErrBadMagic, ErrPageSize, ErrUnsupportedVersion.	Table-driven unit tests cover every invalid permutation plus a golden header file.	
Dynamic page size	Store page size from header; all offset math reuses it. If a legal header change occurs later (e.g. via PRAGMA page_size=) driver re-reads and adapts.	Integration test changes page size, re-opens database, confirms new size honoured.
DSN parsing	Grammar: file:/abs/path.db?mode=rwc&cache=private&_journal_mode=DELETE&_busy_timeout=5000&_page_size=4096.	
• Reject unknown keys (fail-closed).		
• Immutable snapshot after Open().		
• String() returns canonical DSN for audit logs.	20 positive/negative unit tests + Go fuzz target for 30 s on parser entry (no panic).	
Virtual File System	Standard VFS: open/create/delete, truncate, byte-range locks (fcntl or LockFileEx), durable Sync, canonical path.	
Sandboxed VFS: whitelist of absolute paths; every call re-checks; denies .., symlinks, Windows \\?\.		
Optional async VFS (Linux): raw io_uring; constructor gracefully falls back if kernel < 5.6.	Concurrency hammer spawns 1 000 goroutines locking/unlocking; sandbox test blocks traversal exploit.	
Pager & ARC cache	• Column-agnostic pager maps (pageID → offset); maintains dirty set, flushes on demand; grows file safely.	
• ARC-style cache (hand-coded) with fixed page cap, O(1) hit/miss, mutex protection.		
• Memory ceiling: RSS never exceeds configured pages.	Benchmarks show constant RAM when cache limit hit; eviction correctness verified with deterministic access trace.	
Concurrency & durability	• Busy-timeout exponential back-off (default 5 s, DSN tunable). After 3 cycles → ErrDatabaseBusy.	
• Journal mode defaults to DELETE; DSN may request WAL (setting persisted, replay logic arrives Phase 3).		
• Optional nightly hook runs PRAGMA integrity_check; and flips internal health=false on failure.	Race detector passes; corruption injection test sets health=false and surfaces via API.	
Observability	• Logs via log.Logger: ISO-8601 ts, level, traceID, component, msg.	
• File rotation: 10 MiB segments, 14-day TTL at %ProgramData%\gosqlite\logs or /var/log/gosqlite/.		
• Metrics: std-lib expvar counters (pager_reads_total, pager_writes_total, cache_hits_total, busy_retries_total).		
• Optional diagnostics server guarded by env-var; exposes /debug/vars, /debug/pprof/*.	Curl verifies counters; log-roll test hits 10 MiB, sees new file created.	
Quality gates	• Unit coverage ≥ 80 %.	
• go vet → zero diagnostics.		
• go test -race green on linux-amd64, darwin-arm64, windows-amd64.		
• 30 s fuzz on header & DSN parsers—no panic or leak.		
• Performance guard: P99 page read < 50 µs at 50 k QPS on dev box.	All executed in CI with plain go tool-chain.	
Packaging	Single static binary (go build -trimpath -ldflags "-s -w").	
On startup prints Git commit + executable SHA-256.		
No CGO, no external code.	file command shows no dynamic-link sections; SHA-256 matches printed hash.	
Documentation	RUNBOOK.md (install, DSN examples, backup, restore, lock troubleshooting)	
ARCHITECTURE.md (ASCII diagrams: VFS stack, pager/cache flow, optional io_uring path)		
CHANGELOG.md (start at v0.1.0)		
SECURITY.md (threat model: traversal, spoofed header, lock denial; mitigation & integrity-check process)	Docs reviewed and merged before tag.	

✅ Phase 1 exit checklist
Open, validate, read, write, flush real database file with no corruption.

Sandboxed VFS blocks crafted path traversal in automated test.

Cache eviction keeps memory bound in 10-minute 50 k QPS read benchmark.

Logs, metrics, diagnostics work and roll correctly.

All quality gates pass in CI (pure Go only).