Phase 5 — Transactions (Hardened, Pure-Go)
Goal: deliver a provably-correct transaction subsystem — supporting WAL, rollback journals, savepoints and robust file-locking — with zero third-party Go code. Formal proofs live in separate spec documents; production binaries remain pure standard-library Go.

5.1 · Formally-Specified Transaction Manager
Deliverable	“Fully functional” definition	Notes on purity
Formal spec (TLA+ or markdown state-machine)	Captures Atomicity, Consistency, Isolation, Durability, Liveness, Safety for both WAL and rollback-journal modes.	Specification & model-checking files live in /spec/…. They are not compiled into the Go binary, so no CGO or extra libs enter the build.
Model-check evidence	README.md in /spec/ lists model-checker runs that reached “No invariant violations found”.	Model-checker toolchain is installed separately on CI runners; artefacts stored as text.
transaction_manager.go	Pure-Go state-machine that mirrors the spec: Begin, Commit, Rollback, crash-recovery entry‐points.	Uses only sync, time, os, io, and Phase-1 VFS. No reflection or unsafe.
WAL implementation	Appends 32-byte headers + page images to *.wal; maintains in-memory index of frames. Concurrent readers map WAL + DB via VFS locks.	WAL code allocs only on frame-buffer growth; steady-state reuses byte-slices.
Rollback-journal implementation	Writes original pages to *-journal before first modification, then sets journal_header_synced = true.	Same buffering rules as WAL; journal file truncated on commit or rolled back on error.
Crash-recovery logic	On open: detect dirty flag → replay WAL or roll back journal; verify page checksums.	Unit tests craft controlled power-fail scenarios using in-repo helpers.

5.2 · Savepoints & Locking
Deliverable	Detail	Validation
Savepoint stack	Slice-based stack records dirty-page list + counters per savepoint; supports SAVEPOINT, ROLLBACK TO, RELEASE. Depth cap = 512 (configurable).	Table-driven tests exercise nested save/rollback combinations.
Lock types	Implements SHARED, RESERVED, PENDING, EXCLUSIVE exactly like SQLite’s documented protocol.	Multi-process test harness (pure Go) forks helper binaries that acquire conflicting locks; verifies ordering.
Locking API in VFS	Adds ObtainLock(level) and PromoteLock(old,new) to VFS file handle; uses fcntl (*nix) or LockFileEx (Windows) via syscall.	Race detector run with 100 goroutines repeatedly promoting/demoting locks.
Deadlock detection	Each Conn holds wait-graph entry; on timeout the manager walks graph (O(n²) worst-case) to detect cycle → abort youngest txn.	Synthetic two-client test induces cycle, expects youngest aborted in < 100 ms.
Formal lock model	Lock state & transitions added to TLA+/markdown spec; model-checker run shows no reachable deadlock state.	Artefact in /spec/locks/….

Observability additions (std-lib only)
expvar counters: wal_frames_written, journal_pages_written, tx_commits, tx_rollbacks, savepoints_total, deadlocks_detected.

Log lines: on every Commit or failed Rollback, emit structured tx_id, duration_ms, mode=WAL|JOURNAL, pages_dirty, err.

Quality-gates (all standard-library commands)
Gate	Threshold / tool
Unit coverage	transaction/ ≥ 85 % (go test -cover)
Race detector	go test -race ./transaction/... green
Vet	go vet ./transaction/... zero diagnostics
Fuzz	30 s per PR on savepoint stack & lock promotion; no panic/leak
Crash-recovery test	Power-fail simulator runs 1 000 randomized interruption points → always recovers

Documentation
TRANSACTIONS.md – WAL vs journal overview, savepoint semantics, lock table, config knobs.

SPEC_README.md – how to run model checker, interpret counter-examples.

CHANGELOG.md – add v0.4.0 entry for transaction subsystem GA.

✅ Phase 5 exit criteria
5 000-iteration power-fail loop shows zero corruption; replay time ≤ 3 s at 1 GiB WAL.

Concurrent 50-writer / 200-reader stress for 30 min shows no deadlock or lost update.

All gates green; binaries still build with nothing but the Go standard library.