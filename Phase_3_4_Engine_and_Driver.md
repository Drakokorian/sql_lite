Phase 3 — Vectorised Execution Engine (VDBE) — Pure-Go, Zero-Dependency
Goal: replace the classic row-oriented interpreter with a batch-oriented, SIMD-optional engine that runs entirely in Go.
No CGO, no C, no external modules — only the Go 1.x standard library and, optionally, hand-written Go/assembly kept inside the repo.

1 · Columnar Storage & Batch Loop
Deliverable	“Fully functional” definition	Validation (std-lib only)
Column-major buffers	Every table or temp result is held as one typed slice per column ([]int64, []float64, []byte…), never as per-row structs.	Unit test checks pointer difference between successive values equals element size.
Batch size control	Default batch = 1 024 rows; CLI/env var sets 256 – 16 384.	Benchmark shows linear runtime vs batch size with no extra allocs.
Execution driver	Outer loop slices columns into batches; dispatches vector opcodes until rows exhausted.	go test -bench VectorLoop shows ≥4 × speed-up vs reference row loop in same repo.

2 · Vector-Aware Opcodes
Category	Vector behaviour
Comparison (EQ, LT, GT, etc.)	Accept two value slices, produce []bool mask; support NULL semantics via parallel []byte validity bitmap.
Arithmetic	Operate element-wise on numeric slices; overflow saturates to SQLite rules.
Filter	Apply []bool mask to all column slices in constant time (copy-forward algorithm).
Hash & merge joins	Hash join: build side hashed into open-address table; probe side processed in batches.
Merge join: two sorted streams advanced with binary-search helper.
Aggregate	Reduction across slice—SUM, COUNT, AVG, MIN, MAX; grouping uses simple in-memory hash table keyed by group cols.
Constant-time comparisons	Equality routines for sensitive data (e.g., passwords) implemented with length-invariant loops (pure Go); build tag timing_test verifies <2 % delta between equal and diff inputs.

3 · SIMD Optional Hot Paths (Still Zero-Dependency)
Implementation: Hand-written .s files (*_avx2.s, *_neon.s) inside repo; build tag nosimd disables.

Dispatch: Tiny init checks cpu.X86.HasAVX2 or cpu.ARM64.HasASIMD (std-lib internal/cpu exported via runtime/cpu in Go 1.22) and swaps function pointers.

Purity: No unsafe package beyond necessary slice→pointer conversions; still part of std-lib.

4 · Memory & Security Guards
Guard	Policy
Memory cap	Estimate = batchRows × rowSize × columnCount. If > configured cap (default 64 MiB) → ErrExecMemory.
Zero-alloc hot path	Scratch buffers in sync.Pool; after warm-up allocs/query ≤5. Verified with go test -bench -benchmem.
Bounds checking	All slice accesses guarded by if idx >= len(slice) { panic("bounds") } in debug builds; strip with -tags prod if desired.
Context polls	Each batch loop starts with select { case <-ctx.Done(): return ctx.Err() default: } for cooperative cancel.

5 · Observability (std-lib only)
Counters via expvar: vectors_total, simd_calls_total, exec_mem_bytes, join_hash_probe_collisions.

Slow-query log: if execution > threshold (default 50 ms) write structured line with query_id, exec_ms, batch_rows, simd=on|off.

pprof remains enabled from Phase 1; vector engine exposes custom profile labels (runtime/pprof.SetGoroutineLabels) for query_id.

6 · Tests & Quality Gates
Gate	Threshold / method (all std-lib)
Unit coverage	vdbe/ packages ≥ 85 % (go test -cover).
Race detector	go test -race ./vdbe/... green (default & nosimd).
go vet	Zero diagnostics.
Benchmark guard	VectorScan over 10 M rows ≥4× faster than row loop; nosimd build ≥2×.
Fuzz	30 s per PR on opcode argument generator; no panic, leak, or race.
Memory limit	Stress test loads 70 MiB estimate → engine returns ErrExecMemory; 60 MiB passes without OOM.

7 · Documentation
VDBE.md — architecture diagram, opcode table, batch loop, SIMD build tags, memory formula.

PERF_GUIDE.md — tuning batch size, memory cap, nosimd flag, slow-query threshold.

CHANGELOG.md — add v0.3.0 entry for vector engine GA.

✅ Phase 3 Exit Checklist
TPC-H Query 1 at 10 GB runs in <40 s (pure Go) and <20 s (simd build) on reference laptop.

Worst-case adversarial query capped by memory guard; returns ErrExecMemory, never crashes.

10-minute 100-goroutine hammer passes race detector; steady RSS < configured cap.

Logs & metrics show vector counts and SIMD flag; slow-query log fires appropriately.

All gates green; tag release v0.3.0.








