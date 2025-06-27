# **Phase 5: Transactions (Hardened)**

**Primary Goal:** To build a mathematically proven, correct, and secure transaction system.

### **Sprint 5.1: Formally Verified Transaction Logic**

**Objective:** To mathematically prove the correctness and robustness of the transaction management system, particularly for Rollback Journal and Write-Ahead Log (WAL) modes.

#### **Component: Transaction Manager (`transaction_manager.go`)**

1.  **Formal Specification Development:**
    *   **Language Selection:** Utilize a rigorous formal specification language (e.g., TLA+, PlusCal, or a custom-designed state-machine language) to precisely define the behavior of the transaction manager.
    *   **Key Properties:** The specification will capture critical properties such as:
        *   **Atomicity:** All operations within a transaction either complete successfully or are entirely rolled back.
        *   **Consistency:** Transactions bring the database from one valid state to another.
        *   **Isolation:** Concurrent transactions do not interfere with each other.
        *   **Durability:** Once a transaction is committed, its changes are permanent, even in the face of system failures.
        *   **Liveness:** Operations eventually complete.
        *   **Safety:** Nothing bad ever happens (e.g., no data corruption, no deadlocks).
    *   **State Machine Modeling:** Model the transaction manager as a state machine, defining all possible states, transitions, and the conditions under which transitions occur. This includes states for active transactions, pending commits, rollback, and recovery.
2.  **Model Checking and Proof:**
    *   **Automated Verification:** Employ automated model checking tools to exhaustively explore the state space of the formal specification. This process will identify potential race conditions, deadlocks, and logical errors that are extremely difficult to find with traditional testing.
    *   **Theorem Proving:** For more complex properties or infinite state spaces, interactive theorem provers may be used to construct mathematical proofs of correctness.
    *   **Refinement Loop:** The process will be iterative: any discrepancies or violations found by the model checker will lead to refinements in the formal specification and, subsequently, in the Go implementation, ensuring a provably correct system.
3.  **WAL and Rollback Journal Implementation:**
    *   **WAL Mode:** Implement the Write-Ahead Log mechanism, where all changes are first written to a separate WAL file before being applied to the main database file. This enables concurrent reads and writes and efficient crash recovery.
    *   **Rollback Journal Mode:** Implement the traditional rollback journal, where original page contents are written to a journal file before modification, allowing for easy rollback.
    *   **Recovery Logic:** Develop robust recovery procedures for both WAL and rollback journal modes to ensure data integrity after crashes or unexpected shutdowns.

### **Sprint 5.2: Savepoints & Locking**

**Objective:** To extend the transaction system with support for nested transactions (savepoints) and implement a robust, formally verified file locking mechanism.

#### **Component: Transaction Engine (`transaction_engine.go`)**

1.  **Savepoint Implementation:**
    *   **Nested Transactions:** Implement support for `SAVEPOINT`, `RELEASE SAVEPOINT`, and `ROLLBACK TO SAVEPOINT` commands. This allows for finer-grained control over transactions, enabling partial rollbacks within a larger transaction.
    *   **Internal Stack:** Manage savepoints using an internal stack, where each savepoint records the necessary state to revert changes up to that point.
    *   **Resource Management:** Ensure proper resource management (e.g., memory, file handles) for each active savepoint.
2.  **File Locking Mechanism:**
    *   **Concurrency Control:** Implement a sophisticated file locking mechanism to ensure proper concurrency control and prevent data corruption in multi-process or multi-threaded environments.
    *   **Lock Types:** Support various lock types (e.g., shared, exclusive, pending) as required by the SQLite concurrency model.
    *   **Platform-Specific Locking:** Utilize platform-specific locking primitives (e.g., `fcntl` on Unix-like systems, `LockFileEx` on Windows) through the VFS interface to ensure atomic and reliable locking.
    *   **Deadlock Prevention/Detection:** Design the locking protocol to prevent deadlocks where possible, and implement mechanisms for deadlock detection and resolution if prevention is not entirely feasible.
3.  **Formal Verification of Locking:** The file locking logic will be integrated into the formal verification model developed in Sprint 5.1. This will ensure that the locking protocol is free from deadlocks, race conditions, and other concurrency-related issues under all possible scenarios.
