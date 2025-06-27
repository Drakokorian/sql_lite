package pkg

import (
	"fmt"
	"sync"
)

// TransactionEngine is responsible for managing savepoints and the file locking mechanism.
// Its design is intended to be formally verifiable for correctness and robustness.
type TransactionEngine struct {
	// Conceptual representation of the database file for locking purposes.
	// In a real system, this would be an abstraction over the VFS.
	dbFile string
	
	// Mutex to protect access to the file locks.
	// In a distributed system, this would be a distributed lock manager.
	mu sync.Mutex

	// Conceptual file locks held by different owners (e.g., transaction IDs).
	// Maps ownerID to a map of LockType to count (for shared locks) or boolean (for exclusive).
	fileLocks map[string]map[LockType]int

	// Conceptual stack for managing savepoints within a transaction.
	// Each savepoint would store the state necessary to revert changes up to that point.
	savepointStacks map[string][]*Savepoint
}

// Savepoint represents a point within a transaction to which changes can be rolled back.
type Savepoint struct {
	Name string
	// State to be restored upon rollback to this savepoint.
	// This would include things like page versions, cursor positions, etc.
}

// NewTransactionEngine creates a new TransactionEngine instance.
func NewTransactionEngine(dbFile string) *TransactionEngine {
	return &TransactionEngine{
		dbFile: dbFile,
		fileLocks: make(map[LockType]int),
		savepointStacks: make(map[string][]*Savepoint),
	}
}

// AcquireLock attempts to acquire a lock of the specified type for the given owner.
// This method is designed to be part of a formally verifiable locking protocol.
func (te *TransactionEngine) AcquireLock(ownerID string, lockType LockType) error {
	te.mu.Lock()
	defer te.mu.Unlock()

	// Conceptual locking logic. In a real system, this would interact with
	// platform-specific locking primitives (e.g., fcntl, LockFileEx) via the VFS.

	ownerLocks, ok := te.fileLocks[ownerID]
	if !ok {
		ownerLocks = make(map[LockType]int)
		te.fileLocks[ownerID] = ownerLocks
	}

	switch lockType {
	case SharedLock:
		// Allow multiple shared locks.
		ownerLocks[SharedLock]++
		fmt.Printf("TransactionEngine: %s acquired SHARED lock. Count: %d\n", ownerID, ownerLocks[SharedLock])
	case ExclusiveLock:
		// Only one exclusive lock allowed, and no shared locks.
		if te.isLockedByOthers(ownerID, ExclusiveLock) || te.hasSharedLocksByOthers(ownerID) {
			return fmt.Errorf("cannot acquire EXCLUSIVE lock: file is locked by others")
		}
		ownerLocks[ExclusiveLock] = 1
		fmt.Printf("TransactionEngine: %s acquired EXCLUSIVE lock.\n", ownerID)
	default:
		return fmt.Errorf("unsupported lock type: %s", lockType)
	}

	return nil
}

// ReleaseLock releases a lock of the specified type for the given owner.
func (te *TransactionEngine) ReleaseLock(ownerID string, lockType LockType) error {
	te.mu.Lock()
	defer te.mu.Unlock()

	ownerLocks, ok := te.fileLocks[ownerID]
	if !ok {
		return fmt.Errorf("owner %s has no locks", ownerID)
	}

	switch lockType {
	case SharedLock:
		if ownerLocks[SharedLock] > 0 {
			ownerLocks[SharedLock]--
			fmt.Printf("TransactionEngine: %s released SHARED lock. Count: %d\n", ownerID, ownerLocks[SharedLock])
		}
	case ExclusiveLock:
		if ownerLocks[ExclusiveLock] > 0 {
			ownerLocks[ExclusiveLock] = 0
			fmt.Printf("TransactionEngine: %s released EXCLUSIVE lock.\n", ownerID)
		}
	default:
		return fmt.Errorf("unsupported lock type: %s", lockType)
	}

	// Clean up if no locks are held by this owner
	if ownerLocks[SharedLock] == 0 && ownerLocks[ExclusiveLock] == 0 {
		delete(te.fileLocks, ownerID)
	}

	return nil
}

// ReleaseAllLocks releases all locks held by the given owner.
func (te *TransactionEngine) ReleaseAllLocks(ownerID string) error {
	te.mu.Lock()
	defer te.mu.Unlock()

	if _, ok := te.fileLocks[ownerID]; ok {
		delete(te.fileLocks, ownerID)
		fmt.Printf("TransactionEngine: %s released all locks.\n", ownerID)
	}
	return nil
}

// isLockedByOthers checks if the file is exclusively locked by another owner.
func (te *TransactionEngine) isLockedByOthers(currentOwner string, lockType LockType) bool {
	for owner, locks := range te.fileLocks {
		if owner != currentOwner && locks[ExclusiveLock] > 0 {
			return true
		}
	}
	return false
}

// hasSharedLocksByOthers checks if the file has shared locks by other owners.
func (te *TransactionEngine) hasSharedLocksByOthers(currentOwner string) bool {
	for owner, locks := range te.fileLocks {
		if owner != currentOwner && locks[SharedLock] > 0 {
			return true
		}
	}
	return false
}

// CreateSavepoint creates a new savepoint for a given transaction.
// This operation would record the current state of the database for partial rollback.
func (te *TransactionEngine) CreateSavepoint(txID, name string) error {
	te.mu.Lock()
	defer te.mu.Unlock()

	savepoint := &Savepoint{Name: name}
	// In a real implementation, 'savepoint' would capture the current state
	// (e.g., current page versions, changes made so far).

	te.savepointStacks[txID] = append(te.savepointStacks[txID], savepoint)
	fmt.Printf("TransactionEngine: Savepoint '%s' created for transaction %s.\n", name, txID)
	return nil
}

// ReleaseSavepoint releases a named savepoint.
// This means changes up to this savepoint become part of the parent transaction.
func (te *TransactionEngine) ReleaseSavepoint(txID, name string) error {
	te.mu.Lock()
	defer te.mu.Unlock()

	stack, ok := te.savepointStacks[txID]
	if !ok || len(stack) == 0 {
		return fmt.Errorf("no savepoints for transaction %s", txID)
	}

	found := false
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i].Name == name {
			// Remove this savepoint and all subsequent ones from the stack.
			te.savepointStacks[txID] = stack[:i]
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("savepoint '%s' not found for transaction %s", name, txID)
	}

	fmt.Printf("TransactionEngine: Savepoint '%s' released for transaction %s.\n", name, txID)
	return nil
}

// RollbackToSavepoint rolls back the transaction to a named savepoint.
// All changes made after the savepoint are undone.
func (te *TransactionEngine) RollbackToSavepoint(txID, name string) error {
	te.mu.Lock()
	defer te.mu.Unlock()

	stack, ok := te.savepointStacks[txID]
	if !ok || len(stack) == 0 {
		return fmt.Errorf("no savepoints for transaction %s", txID)
	}

	foundIndex := -1
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i].Name == name {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return fmt.Errorf("savepoint '%s' not found for transaction %s", name, txID)
	}

	// Conceptual rollback: undo changes from the current state back to the savepoint.
	// This would involve using the WAL or Rollback Journal to revert page changes.
	fmt.Printf("TransactionEngine: Rolling back transaction %s to savepoint '%s'.\n", txID, name)

	// Discard savepoints above the target savepoint.
	te.savepointStacks[txID] = stack[:foundIndex+1]

	return nil
}
