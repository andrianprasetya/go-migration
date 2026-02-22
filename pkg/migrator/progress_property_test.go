package migrator

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// Feature: library-improvements, Property 7: Progress events have correct indices and totals
// **Validates: Requirements 4.1, 4.2**

func TestPropertyProgress_UpEventsHaveCorrectIndicesAndTotals(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		count := rapid.IntRange(1, 10).Draw(t, "migrationCount")
		baseTS := rapid.Int64Range(20200101000000, 20291231235959).Draw(t, "baseTS")

		// Generate unique sorted migration names.
		names := make([]string, count)
		for i := 0; i < count; i++ {
			names[i] = fmt.Sprintf("%014d_migration_%d", baseTS+int64(i)*100, i)
		}

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// Collect progress events.
		var events []ProgressEvent
		m := New(db, WithGrammar(&mockGrammar{}), WithProgress(func(e ProgressEvent) {
			events = append(events, e)
		}))

		for _, name := range names {
			require.NoError(t, m.Register(name, &noopMigration{}))
		}

		// Set up mock expectations for Up(): no applied migrations.
		expectEnsureTable(mock)
		expectGetApplied(mock, nil)
		expectMaxBatch(mock, 0)
		for _, name := range names {
			expectMigrationTx(mock)
			expectRecord(mock, name, 1)
		}

		err = m.Up()
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())

		// Verify progress events.
		require.Len(t, events, count, "Up should produce exactly N progress events")
		for i, ev := range events {
			assert.Equal(t, i, ev.Index, "event[%d] Index", i)
			assert.Equal(t, count, ev.Total, "event[%d] Total", i)
			assert.Equal(t, names[i], ev.Name, "event[%d] Name", i)
			assert.Equal(t, "up", ev.Direction, "event[%d] Direction", i)
		}
	})
}

func TestPropertyProgress_ResetEventsHaveCorrectIndicesAndTotals(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		count := rapid.IntRange(1, 10).Draw(t, "migrationCount")
		baseTS := rapid.Int64Range(20200101000000, 20291231235959).Draw(t, "baseTS")

		names := make([]string, count)
		for i := 0; i < count; i++ {
			names[i] = fmt.Sprintf("%014d_migration_%d", baseTS+int64(i)*100, i)
		}

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		var events []ProgressEvent
		m := New(db, WithGrammar(&mockGrammar{}), WithProgress(func(e ProgressEvent) {
			events = append(events, e)
		}))

		for _, name := range names {
			require.NoError(t, m.Register(name, &noopMigration{}))
		}

		// Build applied records (all migrations applied, ascending order).
		applied := make([]MigrationRecord, count)
		for i, name := range names {
			applied[i] = MigrationRecord{Name: name, Batch: 1}
		}

		// Set up mock expectations for Reset().
		expectEnsureTable(mock)
		expectGetApplied(mock, applied)

		// Reset reverses applied, so Down runs in reverse order.
		for i := count - 1; i >= 0; i-- {
			expectMigrationTx(mock)
			expectRemove(mock, names[i])
		}

		err = m.Reset()
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())

		// Verify progress events.
		require.Len(t, events, count, "Reset should produce exactly N progress events")
		for i, ev := range events {
			assert.Equal(t, i, ev.Index, "event[%d] Index", i)
			assert.Equal(t, count, ev.Total, "event[%d] Total", i)
			// Reset reverses the applied list, so event names are in reverse order.
			assert.Equal(t, names[count-1-i], ev.Name, "event[%d] Name", i)
			assert.Equal(t, "down", ev.Direction, "event[%d] Direction", i)
		}
	})
}
