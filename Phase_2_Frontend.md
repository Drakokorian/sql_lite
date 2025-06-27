Phase 2 — SQL Front-End (Pure-Go, Zero-Dependency)
Objective: add a secure, resource-bounded SQL language layer on top of the Phase 1 storage engine using only the Go standard library and OS sys-calls—no CGO, no third-party code, no reflection.

1 · Tokenizer — “Zero-Trust” Lexical Layer
Deliverable	“Fully functional” definition	Self-tests (std-lib only)
Full dialect coverage	Recognises every token required for SQLite 3 grammar: keywords, identifiers, numeric & string literals, bind parameters (?, :name, @name, $name), operators, comments.	500-query golden corpus in testdata/ compared byte-for-byte to expected token stream.
Hard input cap	Rejects raw query > 256 KiB → ErrQueryTooLarge. Limit configurable via SetMaxQueryBytes.	Unit test feeds 300 KiB string, asserts error in <1 µs.
Character whitelist	Control bytes < 0x20 (except TAB/LF/CR) illegal; error includes byte offset and line/column.	Table-driven test enumerates all 33 disallowed bytes.
Streaming scan	One pass over []byte; constant-size look-ahead; heap allocations 0 for queries ≤ 4 KiB.	go test -bench Tokenize -benchmem shows 0 allocs for 4 KiB input.
Precise error struct	type LexError struct { Code, ByteOff, Line, Col int; Msg string } — callers get exact position for IDE jump-to-error.	Unit tests assert every field for malformed inputs.

2 · Recursive-Descent Parser
Capability	Hardened rules
AST construction	Handles SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, read-only PRAGMA, EXPLAIN, CTEs, window functions. Each node holds source-range (start, end byte).
Semantic checks	Function arity; aggregate vs GROUP BY; unique column aliases; optional “deterministic-functions-only” mode.
Resource caps	Configurable: ≤ 16 joins, ≤ 8 CTEs, expression depth ≤ 25, token count ≤ 1000, AST nodes ≤ 20 000. Exceed → ErrLimitExceeded.
Context cancellation	Parse(ctx, bytes) aborts quickly (<1 ms) on ctx.Done().
Deterministic fingerprint	Identical input bytes → identical JSON serialisation → SHA-256 fingerprint (implemented with crypto/sha256) for plan cache.

3 · Error Model
Stable codes: E_LEX, E_SYNTAX, E_SEMANTIC, E_LIMIT, E_TOO_LARGE.

Errors implement interface{ Error() string; Code() int; Offset() int }.

Minimal recovery: after fatal token, parser skips to next semicolon so batch files surface multiple errors at once.

4 · Coverage-Guided Fuzzing (Go 1.22 built-in)
Element	Details (no external tools)
Harness	parser_fuzz_test.go: func FuzzParse(f *testing.F) converts []byte→string, runs Tokenize→Parse with 256 KiB cap; ignores context.Canceled.
Corpus	Seeds: ANSI tests, SQLite edge-case queries, pathologically deep parentheses, long identifiers. Stored under testdata/corpus/.
CI budget	30 s per PR (go test -fuzztime=30s ./...), nightly 15 min with -fuzztime=15m.
Crash artifacts	On panic, failing input is written to testdata/crash_*.sql via os.WriteFile.
Memory leak guard	Harness snapshots runtime.MemStats every 10 000 iterations; > 5 % growth fails fuzz job.

5 · Observability (std-lib only)
Logs via log.Logger (already initialised in Phase 1): add fields phase=tokenize|parse, query_len, ast_nodes, duration_ms.

Metrics via expvar:

counters sql_parse_total, sql_parse_errors_total

histogram buckets sql_parse_latency_ms_{1,4,16,64,256} (manual slice counters)

Slow-parse log: if parse > configured threshold (default 10 ms) emit "slow_sql_parse" entry with first 1 KiB of UTF-8-escaped query.

6 · Quality Gates
Gate	Threshold & tool (std-lib)
Unit coverage	Tokeniser + parser ≥ 85 % (go test -cover)
Race detector	go test -race ./... green on linux-amd64, darwin-arm64, windows-amd64
Vet	go vet ./... zero diagnostics
Fuzz	30 s run on PR; no panic, leak or race
Bench guard	go test -bench ParseLong -benchtime=10x -benchmem must not regress > 15 % time or allocs vs main

All gates execute with the plain Go tool-chain—no third-party linters or analyzers.

7 · Documentation Updates
PARSER.md – grammar table, AST node glossary, limit knobs, fingerprint rationale.

FUZZING.md – how to run/debug fuzz locally, corpus layout, crash triage SOP.

SECURITY.md – new section on SQL input threat vectors and enforced limits.

CHANGELOG.md – add v0.2.0 entry for SQL front-end GA.

✅ Phase 2 Exit Checklist
Parses ≥ 99 % of public SQLite regression queries (imported as plain text in repo) without error; malformed inputs return correct codes & offsets.

Worst-case adversarial input consumes ≤ 64 MiB RAM, aborts ≤ 250 ms with E_LIMIT.

24-hour continuous fuzz (nightly) finishes without crash, leak, or race.

Logs and expvar metrics appear correctly for 10 000 parses.

All quality gates green; tag release v0.2.0.








