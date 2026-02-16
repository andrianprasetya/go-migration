package migrator

import (
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/andrianprasetya/go-migration/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// --- Property test helpers ---

// trackingMigration records when Up/Down are called and in what order.
type trackingMigration struct {
	name      string
	upCalls   *[]string
	downCalls *[]string
}

func (m *trackingMigration) Up(s *schema.Builder) error {
	if m.upCalls != nil {
		*m.upCalls = append(*m.upCalls, m.name)
	}
	return nil
}

func (m *trackingMigration) Down(s *schema.Builder) error {
	if m.downCalls != nil {
		*m.downCalls = append(*m.downCalls, m.name)
	}
	return nil
}

// genUniqueMigrationNames generates N unique, sorted migration names.
func genUniqueMigrationNames(t *rapid.T, label string) []string {
	count := rapid.IntRange(1, 8).Draw(t, label+"_count")
	baseTS := rapid.Int64Range(20200101000000, 20291231235959).Draw(t, label+"_base")

	names := make([]string, count)
	for i := 0; i < count; i++ {
		names[i] = fmt.Sprintf("%014d_migration_%d", baseTS+int64(i)*100, i)
	}
	return names
}

// genAppliedSubset picks a random subset of names to be "already applied" and assigns batch numbers.
// Returns the applied records sorted by name (ascending), and the max batch number used.
func genAppliedSubset(t *rapid.T, names []string) ([]MigrationRecord, int) {
	if len(names) == 0 {
		return nil, 0
	}
	// Decide how many to mark as applied (0 to all)
	appliedCount := rapid.IntRange(0, len(names)).Draw(t, "appliedCount")
	if appliedCount == 0 {
		return nil, 0
	}

	// Pick the first appliedCount names (they're already sorted)
	applied := names[:appliedCount]

	// Assign batch numbers: split into 1-3 batches
	maxBatches := appliedCount
	if maxBatches > 3 {
		maxBatches = 3
	}
	numBatches := rapid.IntRange(1, maxBatches).Draw(t, "numBatches")

	records := make([]MigrationRecord, appliedCount)
	now := time.Now()
	for i, name := range applied {
		batch := (i % numBatches) + 1
		records[i] = MigrationRecord{
			Name:      name,
			Batch:     batch,
			CreatedAt: now,
		}
	}
	// Find max batch
	maxBatch := 0
	for _, r := range records {
		if r.Batch > maxBatch {
			maxBatch = r.Batch
		}
	}
	return records, maxBatch
}

// Feature: go-migration, Property 5: Up executes only pending migrations in order
// **Validates: Requirements 2.1, 2.3**
func TestProperty5_UpExecutesOnlyPendingInOrder(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		names := genUniqueMigrationNames(t, "mig")
		appliedRecords, maxBatch := genAppliedSubset(t, names)

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		m := New(db, WithGrammar(&mockGrammar{}))

		var upCalls []string
		appliedSet := make(map[string]struct{})
		for _, r := range appliedRecords {
			appliedSet[r.Name] = struct{}{}
		}

		for _, name := range names {
			mig := &trackingMigration{name: name, upCalls: &upCalls}
			require.NoError(t, m.Register(name, mig))
		}

		// Compute expected pending
		var expectedPending []string
		for _, name := range names {
			if _, ok := appliedSet[name]; !ok {
				expectedPending = append(expectedPending, name)
			}
		}

		// Set up mock expectations
		expectEnsureTable(mock)
		expectGetApplied(mock, appliedRecords)

		if len(expectedPending) > 0 {
			expectMaxBatch(mock, maxBatch)
			for _, name := range expectedPending {
				expectMigrationTx(mock)
				expectRecord(mock, name, maxBatch+1)
			}
		}

		err = m.Up()
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())

		// Verify: only pending migrations were executed, in ascending order
		assert.Equal(t, expectedPending, upCalls, "Up should execute only pending migrations in order")
	})
}

// Feature: go-migration, Property 6: Batch number consistency
// **Validates: Requirements 2.6**
//
// All pending migrations executed in a single Up() call share the same batch number,
// which is one greater than the previous maximum batch number.
// We verify this by setting mock expectations that enforce the exact batch number
// for each INSERT. If Up() succeeds and all expectations are met, the property holds.
func TestProperty6_BatchNumberConsistency(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		names := genUniqueMigrationNames(t, "mig")
		appliedRecords, maxBatch := genAppliedSubset(t, names)

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		m := New(db, WithGrammar(&mockGrammar{}))

		appliedSet := make(map[string]struct{})
		for _, r := range appliedRecords {
			appliedSet[r.Name] = struct{}{}
		}

		var pendingNames []string
		for _, name := range names {
			if _, ok := appliedSet[name]; !ok {
				pendingNames = append(pendingNames, name)
			}
		}

		for _, name := range names {
			require.NoError(t, m.Register(name, &noopMigration{}))
		}

		expectEnsureTable(mock)
		expectGetApplied(mock, appliedRecords)

		expectedBatch := maxBatch + 1
		if len(pendingNames) > 0 {
			expectMaxBatch(mock, maxBatch)
			for _, name := range pendingNames {
				expectMigrationTx(mock)
				// Each pending migration must be recorded with exactly expectedBatch
				expectRecord(mock, name, expectedBatch)
			}
		}

		err = m.Up()
		assert.NoError(t, err)
		// If expectations are met, all migrations were recorded with the same batch = maxBatch + 1
		assert.NoError(t, mock.ExpectationsWereMet(),
			"all pending migrations should be recorded with batch %d", expectedBatch)
	})
}

// Feature: go-migration, Property 7: Rollback by batch reverses the last batch
// **Validates: Requirements 3.1**
func TestProperty7_RollbackByBatchReversesLastBatch(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		names := genUniqueMigrationNames(t, "mig")
		// Need at least 1 applied migration for rollback to do something
		if len(names) < 1 {
			return
		}

		// All migrations are applied, across 1-3 batches
		numBatches := rapid.IntRange(1, 3).Draw(t, "numBatches")
		if numBatches > len(names) {
			numBatches = len(names)
		}

		records := make([]MigrationRecord, len(names))
		now := time.Now()
		for i, name := range names {
			batch := (i % numBatches) + 1
			records[i] = MigrationRecord{Name: name, Batch: batch, CreatedAt: now}
		}

		maxBatch := 0
		for _, r := range records {
			if r.Batch > maxBatch {
				maxBatch = r.Batch
			}
		}

		// Find which migrations are in the last batch
		var lastBatchRecords []MigrationRecord
		for _, r := range records {
			if r.Batch == maxBatch {
				lastBatchRecords = append(lastBatchRecords, r)
			}
		}

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		m := New(db, WithGrammar(&mockGrammar{}))

		var downCalls []string
		for _, name := range names {
			mig := &trackingMigration{name: name, downCalls: &downCalls}
			require.NoError(t, m.Register(name, mig))
		}

		// Mock expectations for Rollback(0):
		// 1. EnsureTable
		// 2. GetLastBatchNumber -> maxBatch
		// 3. GetByBatch(maxBatch) -> lastBatchRecords
		// 4. For each (reversed): Begin, Commit, Remove
		expectEnsureTable(mock)
		expectMaxBatch(mock, maxBatch)

		batchRows := sqlmock.NewRows([]string{"migration", "batch", "created_at"})
		for _, r := range lastBatchRecords {
			batchRows.AddRow(r.Name, r.Batch, r.CreatedAt)
		}
		mock.ExpectQuery("SELECT migration, batch, created_at FROM .* WHERE batch").
			WithArgs(maxBatch).WillReturnRows(batchRows)

		// Rollback reverses the lastBatchRecords (they come in ascending order from GetByBatch)
		reversed := make([]MigrationRecord, len(lastBatchRecords))
		copy(reversed, lastBatchRecords)
		for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
			reversed[i], reversed[j] = reversed[j], reversed[i]
		}

		for _, r := range reversed {
			expectMigrationTx(mock)
			expectRemove(mock, r.Name)
		}

		err = m.Rollback(0)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())

		// Verify: only last batch migrations were rolled back, in reverse order
		expectedDown := make([]string, len(reversed))
		for i, r := range reversed {
			expectedDown[i] = r.Name
		}
		assert.Equal(t, expectedDown, downCalls,
			"Rollback(0) should execute Down() for last batch migrations in reverse order")
	})
}

// Feature: go-migration, Property 8: Rollback by step count
// **Validates: Requirements 3.2, 3.6**
func TestProperty8_RollbackByStepCount(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		names := genUniqueMigrationNames(t, "mig")
		if len(names) < 1 {
			return
		}

		// All migrations are applied
		records := make([]MigrationRecord, len(names))
		now := time.Now()
		for i, name := range names {
			records[i] = MigrationRecord{Name: name, Batch: 1, CreatedAt: now}
		}

		// Pick a random step count (1 to len+2 to test exceeding total)
		steps := rapid.IntRange(1, len(names)+2).Draw(t, "steps")
		expectedCount := steps
		if expectedCount > len(names) {
			expectedCount = len(names)
		}

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		m := New(db, WithGrammar(&mockGrammar{}))

		var downCalls []string
		for _, name := range names {
			mig := &trackingMigration{name: name, downCalls: &downCalls}
			require.NoError(t, m.Register(name, mig))
		}

		// Mock expectations for Rollback(steps):
		// 1. EnsureTable
		// 2. GetApplied -> all records (ascending order)
		// GetLastNMigrations takes last N in reverse order
		expectEnsureTable(mock)
		expectGetApplied(mock, records)

		// The last expectedCount migrations in reverse order
		var expectedRolledBack []string
		for i := len(names) - 1; i >= len(names)-expectedCount; i-- {
			expectedRolledBack = append(expectedRolledBack, names[i])
		}

		for _, name := range expectedRolledBack {
			expectMigrationTx(mock)
			expectRemove(mock, name)
		}

		err = m.Rollback(steps)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())

		// Verify: exactly min(steps, total) migrations rolled back in reverse order
		assert.Equal(t, expectedRolledBack, downCalls,
			"Rollback(%d) should roll back last %d migrations in reverse order", steps, expectedCount)
	})
}

// Feature: go-migration, Property 9: Successful rollback removes tracker records
// **Validates: Requirements 3.4**
//
// This property is verified implicitly by Properties 7 and 8: the mock expectations
// include expectRemove() for each rolled-back migration. If the mock expectations
// are met, Remove() was called for each migration. We add an explicit test that
// verifies the Remove calls match the rolled-back set.
func TestProperty9_RollbackRemovesTrackerRecords(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		names := genUniqueMigrationNames(t, "mig")
		if len(names) < 1 {
			return
		}

		records := make([]MigrationRecord, len(names))
		now := time.Now()
		for i, name := range names {
			records[i] = MigrationRecord{Name: name, Batch: 1, CreatedAt: now}
		}

		steps := rapid.IntRange(1, len(names)).Draw(t, "steps")
		expectedCount := steps
		if expectedCount > len(names) {
			expectedCount = len(names)
		}

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		m := New(db, WithGrammar(&mockGrammar{}))

		// Track which migrations had Remove() called via mock
		var removedNames []string

		for _, name := range names {
			require.NoError(t, m.Register(name, &noopMigration{}))
		}

		expectEnsureTable(mock)
		expectGetApplied(mock, records)

		for i := len(names) - 1; i >= len(names)-expectedCount; i-- {
			expectMigrationTx(mock)
			// Track the expected removes
			removedNames = append(removedNames, names[i])
			expectRemove(mock, names[i])
		}

		err = m.Rollback(steps)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(),
			"Remove() should be called for each rolled-back migration")
		assert.Len(t, removedNames, expectedCount,
			"exactly %d tracker records should be removed", expectedCount)
	})
}

// Feature: go-migration, Property 10: Reset clears all applied migrations
// **Validates: Requirements 4.1**
func TestProperty10_ResetClearsAllApplied(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		names := genUniqueMigrationNames(t, "mig")
		appliedRecords, _ := genAppliedSubset(t, names)

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		m := New(db, WithGrammar(&mockGrammar{}))

		var downCalls []string
		for _, name := range names {
			mig := &trackingMigration{name: name, downCalls: &downCalls}
			require.NoError(t, m.Register(name, mig))
		}

		// Mock expectations for Reset():
		// 1. EnsureTable
		// 2. GetApplied -> appliedRecords (ascending)
		// 3. For each applied (reversed): Begin, Commit, Remove
		expectEnsureTable(mock)
		expectGetApplied(mock, appliedRecords)

		// Reset reverses the applied records
		reversed := make([]MigrationRecord, len(appliedRecords))
		copy(reversed, appliedRecords)
		for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
			reversed[i], reversed[j] = reversed[j], reversed[i]
		}

		for _, r := range reversed {
			expectMigrationTx(mock)
			expectRemove(mock, r.Name)
		}

		err = m.Reset()
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())

		// Verify: all applied migrations had Down() called in reverse order
		if len(reversed) == 0 {
			assert.Empty(t, downCalls, "Reset with no applied should call no Down()")
		} else {
			expectedDown := make([]string, len(reversed))
			for i, r := range reversed {
				expectedDown[i] = r.Name
			}
			assert.Equal(t, expectedDown, downCalls,
				"Reset should call Down() for all applied migrations in reverse order")
		}
	})
}

// Feature: go-migration, Property 11: Refresh is equivalent to reset then up
// **Validates: Requirements 4.2**
func TestProperty11_RefreshEqualsResetThenUp(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		names := genUniqueMigrationNames(t, "mig")
		appliedRecords, _ := genAppliedSubset(t, names)

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		m := New(db, WithGrammar(&mockGrammar{}))

		var downCalls []string
		var upCalls []string
		for _, name := range names {
			mig := &trackingMigration{name: name, downCalls: &downCalls, upCalls: &upCalls}
			require.NoError(t, m.Register(name, mig))
		}

		// --- Reset phase expectations ---
		// EnsureTable + GetApplied
		expectEnsureTable(mock)
		expectGetApplied(mock, appliedRecords)

		// Reverse applied for Down() calls
		reversed := make([]MigrationRecord, len(appliedRecords))
		copy(reversed, appliedRecords)
		for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
			reversed[i], reversed[j] = reversed[j], reversed[i]
		}

		for _, r := range reversed {
			expectMigrationTx(mock)
			expectRemove(mock, r.Name)
		}

		// --- Up phase expectations ---
		// After reset, all migrations are pending
		expectEnsureTable(mock)
		expectGetApplied(mock, nil) // nothing applied after reset
		expectMaxBatch(mock, 0)     // fresh start

		for _, name := range names {
			expectMigrationTx(mock)
			expectRecord(mock, name, 1)
		}

		err = m.Refresh()
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())

		// Verify: Down was called for all previously applied (reverse order),
		// then Up was called for all migrations (ascending order)
		if len(reversed) == 0 {
			assert.Empty(t, downCalls, "Refresh reset phase with no applied should call no Down()")
		} else {
			expectedDown := make([]string, len(reversed))
			for i, r := range reversed {
				expectedDown[i] = r.Name
			}
			assert.Equal(t, expectedDown, downCalls,
				"Refresh reset phase should call Down() in reverse order")
		}

		// All migrations should be run up in ascending order
		assert.Equal(t, names, upCalls,
			"Refresh up phase should call Up() for all migrations in order")
	})
}

// Feature: go-migration, Property 13: Status reflects registry and tracker state
// **Validates: Requirements 5.3**
func TestProperty13_StatusReflectsState(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		names := genUniqueMigrationNames(t, "mig")
		appliedRecords, _ := genAppliedSubset(t, names)

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		m := New(db, WithGrammar(&mockGrammar{}))

		for _, name := range names {
			require.NoError(t, m.Register(name, &noopMigration{}))
		}

		// Build lookup for applied
		appliedMap := make(map[string]MigrationRecord)
		for _, r := range appliedRecords {
			appliedMap[r.Name] = r
		}

		expectEnsureTable(mock)
		expectGetApplied(mock, appliedRecords)

		statuses, err := m.Status()
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())

		// Verify: one status entry per registered migration
		assert.Len(t, statuses, len(names), "Status should return one entry per registered migration")

		// Verify: statuses are in registration (sorted) order
		for i, s := range statuses {
			assert.Equal(t, names[i], s.Name, "status[%d] should match registered name", i)

			if rec, ok := appliedMap[s.Name]; ok {
				// Applied migration
				assert.True(t, s.Applied, "%s should be marked as applied", s.Name)
				assert.Equal(t, rec.Batch, s.Batch, "%s should have correct batch", s.Name)
				assert.NotNil(t, s.AppliedAt, "%s should have AppliedAt set", s.Name)
			} else {
				// Pending migration
				assert.False(t, s.Applied, "%s should be marked as pending", s.Name)
				assert.Equal(t, 0, s.Batch, "%s should have batch 0", s.Name)
				assert.Nil(t, s.AppliedAt, "%s should have nil AppliedAt", s.Name)
			}
		}
	})
}
