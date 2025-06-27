Phase 6 — Optimization via Just-In-Time (JIT) Compilation — Pure Go
Goal: add a JIT layer that turns hot VDBE batches into native code without CGO, without third-party modules, and without embedding any external assembler.
We rely only on the Go tool-chain, the standard library (unsafe, syscall, runtime, internal/cpu) and hand-crafted byte-emitters kept in-tree.

6.1 · JIT Compiler
Deliverable	“Fully functional” definition	Purity guard-rails
Hot-spot profiler	Per-statement counters track executions and aggregate nanoseconds. When execs ≥ 100 and cum µs ≥ 50 000, the statement is tagged hot. Thresholds are tunable via SetJITHotCounts.	Uses sync/atomic for counters; no external profiler.
Code emitter (jit_compiler.go)	Converts a limited, high-payoff subset of VDBE opcodes (Column, Integer, Real, String, EQ, LT, GT, Add, Sub, Mul, Div, Filter, Project) into raw x86-64 or AArch64 machine code.	Two small byte-slices per arch (emit_x86.go, emit_arm64.go) contain hand-written opcode templates; guarded by //go:build amd64 or arm64.
Executable buffer	Allocates RWX memory with syscall.Mmap, writes machine code, then executes via func(ptr uintptr) uintptr converted with unsafe.Pointer → uintptr → *func().	mmap and mprotect calls are in std-lib syscall; no CGO.
Security rules	• RW→RX toggle after write (mprotect).	
• Code buffer length ≤ 64 KiB; overflows abort.		
• Emitters never copy untrusted SQL text into code space.	Fuzz test mutates VDBE bytecode; any self-modifying attempt triggers bounds panic.	
Adaptive re-opt	If runtime branch counters show >15 % mis-prediction, re-emit code with narrowed type assumptions (e.g., switch to integer fast-path).	Re-opt limited to 3 generations; guard prevents churn.
JIT cache	LRU keyed by fingerprint (from Phase 2). Capacity default 128 compiled plans or 16 MiB, whichever first; eviction purges oldest.	Implemented with ring list + map (all std-lib containers).
Invalidation hooks	On schema change or ANALYZE, planner bumps a global epoch; cache entry keeps epochSeen; mismatch → discard.	Unit test alters schema; cached code invalidates and recompiles.

6.2 · Integration with Planner & VDBE
Plan flow: parser → planner → VDBE bytecode → JIT check:
cold → interpret; hot → compile once, replace entrypoint.

Fallback safety: any panic in compiled code sets atomic flag; engine reverts to interpreter and logs jit_disabled_panic.

SIMD synergy: emitted code re-uses the Phase 3 SIMD helpers for vector ops when cpu.X86.HasAVX2 / cpu.ARM64.HasASIMD is true.

Observability
expvar counters: jit_compiles_total, jit_hits_total, jit_evictions_total, jit_reopts_total, jit_fallback_panic.

Slow-compile log: if compile time > 10 ms log query_id, bytes, gen_ms.

/debug/jit HTTP handler (opt-in): dumps current cache keys and sizes.

Tests & Quality Gates
Gate	Threshold / tool (std-lib)
Unit coverage	jit/ ≥ 80 % (go test -cover)
Race detector	go test -race ./jit/... (interpreter fallback in tests)
Vet	go vet ./jit/... zero diagnostics
Micro-bench	Hot SELECT with int filter runs ≥ 10 × faster w/ JIT than interpreter
Fuzz	30 s fuzz on bytecode → compile-and-run cycle; no panic, no segfault
RWX policy	Unit test ensures code pages are PROT_EXEC only after emission

All gates executed with go commands alone; no CGO, no third-party tools.

Documentation
JIT.md – hot-spot thresholds, architecture diagrams, emitter tables, security notes on RWX.

PERF_GUIDE.md – when to raise/lower compile thresholds, cache size tuning.

CHANGELOG.md – add v0.5.0 entry for JIT feature.

✅ Phase 6 Exit Checklist
Representative TPC-H workload shows ≥ 2× throughput improvement with JIT on reference laptop.

24-hour fuzz & stress yields no crashes; any runtime panic automatically disables JIT and logs.

Memory cap respected (16 MiB code).

All CI gates green; binary remains pure Go only.








