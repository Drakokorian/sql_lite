# **Project: gosqlite - A Hyper-Scale, Hardened SQLite Driver for Go**

## **Mission**
To build a pure Go SQLite driver engineered for mission-critical, high-performance environments. The driver will be a secure, portable, and dependency-free component suitable for use in massively parallel, hyper-scale architectures.

---

## **Governing Pillars**
0. Must be built like an enterprise-grade product—no exceptions.

1.  **Hyper-Scale Architecture Support:**
    *   The driver is designed as a component for sharded, distributed systems.
    *   Features first-class support for WAL-based replication, enabling continuous data streaming and recovery.

2.  **Extreme Performance Optimizations:**
    *   **Asynchronous I/O:** Leverages advanced kernel interfaces on Linux for maximum throughput.
    *   **Vectorized Execution (SIMD):** The VDBE operates on batches of data for CPU-level parallelism.
    *   **Just-In-Time (JIT) Compilation:** Hot queries are compiled to native machine code.
    *   **Zero-Allocation Design:** A strict focus on minimizing memory allocations in hot paths.

3.  **Military-Grade Security & Reliability:**
    *   **Formal Verification:** Transaction logic will be mathematically proven correct using rigorous formal methods.
    *   **Fuzz Testing:** A continuous fuzzing suite will run against the parser to prevent security vulnerabilities.
    *   **Sandboxing:** A swappable, sandboxed VFS enforces the Principle of Least Privilege.
    *   **Secure Supply Chain:** Automated SBOM generation and vulnerability scanning.

4.  **Mission-Critical Delivery:**
    *   Every milestone is delivered as complete, production-ready code. **No stubs, no shortcuts, no deferred work.**

---

## **Coding Standards**

The code-base adheres to a concise but uncompromising set of rules:

1. **Logging** – All critical operations and errors are logged in UTC to an external, append-only text file.  Rotation keeps 14 days of history, with each file capped at 10 MiB.
2. **Mission-Critical Delivery** – Every milestone ships as production-ready, fully tested code.  No stubs, scaffolding, or deferred work are permitted.
3. **Compatibility** – The driver must work with every SQLite file format version from 2008 forward.
4. **Quality Gates**
   • Each exported symbol carries Go doc comments.
   • `go vet`, `go test -race`, and 80 %+ line coverage are mandatory.
5. **Design** – Functions and packages obey the Single Responsibility Principle.  External input is validated; errors wrap context (`fmt.Errorf("component: %w", err)`).
6. **Security** – No third-party runtime dependencies are allowed; the project builds with the Go standard library alone.  Sandboxed VFS blocks path traversal and privilege escalation.
7. **Performance** – Hot paths avoid heap allocations and are benchmarked to ensure < 50 µs P99 page reads at 50 k QPS on commodity hardware.

These standards are non-negotiable and apply to every pull request.