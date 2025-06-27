# **Phases 3 & 4: VDBE & Driver (Hardened)**

**Primary Goal:** To build a hyper-performance, vectorized, and secure execution engine.

### **Sprint 3.1: Vectorized (SIMD) VDBE**

**Objective:** To implement a Virtual Database Engine (VDBE) that processes data in a vectorized manner, leveraging CPU-level parallelism for extreme performance.

#### **Component: VDBE Core (`vdbe.go`)**

1.  **Vectorized Execution Model:** The VDBE will fundamentally shift from a row-at-a-time processing model to a vectorized model. This means that instead of fetching and operating on individual rows, the VDBE will operate on batches (vectors) of data.
    *   **Columnar Data Representation:** Internally, data within the VDBE's execution context (e.g., temporary tables, intermediate results) will be stored in a columnar format. This layout is highly efficient for vectorized operations, as it allows for contiguous memory access for a single column across multiple rows.
    *   **Batch Processing:** Each VDBE opcode will be designed to accept and produce vectors of values. For example, an `OP_EQ` (equality comparison) opcode would take two input vectors and produce a boolean result vector, indicating equality for each pair of elements in the batch.
    *   **CPU-Level Parallelism (SIMD):** The implementation will strategically utilize CPU Single Instruction, Multiple Data (SIMD) instructions where applicable. This will involve:
        *   **Go Assembly/Intrinsics:** For critical hot paths (e.g., string comparisons, numeric operations, hashing), direct Go assembly or compiler intrinsics (if available and stable) will be employed to emit SIMD instructions. This ensures that the CPU performs the same operation on multiple data elements concurrently, maximizing throughput.
        *   **Optimized Data Structures:** Data structures will be aligned to cache lines and designed to minimize cache misses, further enhancing SIMD effectiveness.
2.  **VDBE Opcodes Redesign:** All existing and new VDBE opcodes will be re-engineered to operate on vectors. This includes:
    *   **Filtering Operations:** `WHERE` clauses will be evaluated by applying vectorized comparison opcodes (e.g., `OP_GT`, `OP_LT`, `OP_EQ`) to entire columns, producing boolean masks.
    *   **Aggregation Functions:** `SUM`, `COUNT`, `AVG` will operate on vectors of numbers, accumulating results efficiently.
    *   **Join Operations:** Hash joins and merge joins will be optimized for vectorized input, leveraging vectorized hashing and comparison.

### **Sprint 3.2: Constant-Time & Zero-Allocation Execution**

**Objective:** To ensure that all VDBE operations are designed for maximum security against side-channel attacks and minimal memory footprint.

#### **Component: Hardened VDBE Opcodes (`vdbe_opcodes_hardened.go`)**

1.  **Zero-Allocation Design:**
    *   **Pre-allocated Buffers:** Minimize dynamic memory allocations during query execution by using pre-allocated buffer pools for intermediate results, string manipulations, and other temporary data.
    *   **Stack-based Allocations:** Favor stack-based allocations for small, short-lived data structures where possible, reducing garbage collector pressure.
    *   **In-place Operations:** Prioritize in-place modifications of data structures to avoid creating new copies.
2.  **Constant-Time Algorithms:**
    *   **Side-Channel Attack Prevention:** Any VDBE opcode that performs comparisons or operations on potentially sensitive user data (e.g., password hashes, encryption keys, authentication tokens) will strictly use constant-time algorithms. This prevents timing side-channel attacks, where an attacker could infer information about secret values by measuring the execution time of operations.
    *   **Implementation:** This involves careful implementation of comparison routines to ensure that their execution time is independent of the input data's values. This may require custom Go implementations or leveraging cryptographic libraries that provide constant-time primitives.
3.  **Input Validation and Bounds Checking:**
    *   **Strict Validation:** Every opcode will perform rigorous validation of its input operands, ensuring they are within expected ranges and types.
    *   **Bounds Checking:** Explicit bounds checking will be performed for all array and slice accesses to prevent out-of-bounds reads or writes, which could lead to crashes or security vulnerabilities.

### **Sprint 4.1: Go `database/sql` Integration**

**Objective:** To provide a standard and idiomatic Go interface for interacting with the `gosqlite` driver, adhering to the `database/sql` package specifications.

#### **Component: Driver Interface (`driver.go`)**

1.  **`database/sql/driver` Implementation:** The `gosqlite` driver will implement the interfaces defined in Go's `database/sql/driver` package. This includes:
    *   **`Driver` Interface:** For opening new database connections.
    *   **`Conn` Interface:** For managing database sessions, preparing statements, and beginning transactions.
    *   **`Stmt` Interface:** For executing prepared statements and handling parameters.
    *   **`Result` Interface:** For retrieving results from `INSERT`, `UPDATE`, `DELETE` operations.
    *   **`Rows` Interface:** For iterating over query results.
2.  **Error Mapping:** The driver will correctly map the detailed internal SQLite error codes and messages to standard Go error types and `database/sql` specific errors. This ensures consistent and predictable error handling for applications using the driver.
3.  **Parameter Handling:** Implement robust handling of query parameters, ensuring proper type conversion and prevention of SQL injection vulnerabilities.
4.  **Connection Pooling:** While `database/sql` handles connection pooling at a higher level, the `gosqlite` driver will ensure its `Conn` implementations are lightweight and efficient to support effective pooling.
5.  **Concurrency Model:** Ensure that the driver's implementation of the `database/sql` interfaces is thread-safe and handles concurrent access to connections and statements correctly, aligning with Go's concurrency model.
6.  **Context Propagation:** Support `context.Context` for cancellation and timeouts in database operations, allowing applications to manage long-running queries and resource usage effectively.
