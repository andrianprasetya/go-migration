package migrator

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/andrianprasetya/go-migration/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Helpers ---

func newTestMigrator(t *testing.T) (*Migrator, *sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	m := New(db, WithGrammar(&mockGrammar{}))
	return m, db, mock
}

// expectEnsureTable sets up the sqlmock expectation for CREATE TABLE IF NOT EXISTS.
func expectEnsureTable(mock sqlmock.Sqlmock) {
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
}

// expectGetApplied sets up a query expectation returning the given migration names/batches.
func expectGetApplied(mock sqlmock.Sqlmock, records []MigrationRecord) {
	rows := sqlmock.NewRows([]string{"migration", "batch", "created_at"})
	for _, r := range records {
		rows.AddRow(r.Name, r.Batch, r.CreatedAt)
	}
	mock.ExpectQuery("SELECT migration, batch, created_at FROM").WillReturnRows(rows)
}

// expectMaxBatch sets up the COALESCE(MAX(batch), 0) query.
func expectMaxBatch(mock sqlmock.Sqlmock, batch int) {
	mock.ExpectQuery("SELECT COALESCE").WillReturnRows(
		sqlmock.NewRows([]string{"max"}).AddRow(batch),
	)
}

// expectRecord sets up an INSERT expectation for recording a migration.
func expectRecord(mock sqlmock.Sqlmock, name string, batch int) {
	mock.ExpectExec("INSERT INTO").WithArgs(name, batch).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

// expectRemove sets up a DELETE expectation for removing a migration record.
func expectRemove(mock sqlmock.Sqlmock, name string) {
	mock.ExpectExec("DELETE FROM").WithArgs(name).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

// expectMigrationTx sets up Begin + Commit expectations for a single migration.
func expectMigrationTx(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectCommit()
}

// --- Up Tests ---

func TestUp_PendingMigrations(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	require.NoError(t, m.Register("20240101000000_create_users", &noopMigration{}))
	require.NoError(t, m.Register("20240102000000_create_posts", &noopMigration{}))

	expectEnsureTable(mock)
	expectGetApplied(mock, nil) // no applied migrations
	expectMaxBatch(mock, 0)     // first batch

	// Migration 1
	expectMigrationTx(mock)
	expectRecord(mock, "20240101000000_create_users", 1)

	// Migration 2
	expectMigrationTx(mock)
	expectRecord(mock, "20240102000000_create_posts", 1)

	err := m.Up()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUp_NoPending(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	require.NoError(t, m.Register("20240101000000_create_users", &noopMigration{}))

	expectEnsureTable(mock)
	expectGetApplied(mock, []MigrationRecord{
		{Name: "20240101000000_create_users", Batch: 1, CreatedAt: time.Now()},
	})

	err := m.Up()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUp_PartiallyApplied(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	require.NoError(t, m.Register("20240101000000_create_users", &noopMigration{}))
	require.NoError(t, m.Register("20240102000000_create_posts", &noopMigration{}))

	expectEnsureTable(mock)
	expectGetApplied(mock, []MigrationRecord{
		{Name: "20240101000000_create_users", Batch: 1, CreatedAt: time.Now()},
	})
	expectMaxBatch(mock, 1) // next batch = 2

	// Only migration 2 should run
	expectMigrationTx(mock)
	expectRecord(mock, "20240102000000_create_posts", 2)

	err := m.Up()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Rollback Tests ---

func TestRollback_ByBatch(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	require.NoError(t, m.Register("20240101000000_create_users", &noopMigration{}))
	require.NoError(t, m.Register("20240102000000_create_posts", &noopMigration{}))

	expectEnsureTable(mock)

	// GetLastBatch: max batch = 1, then GetByBatch(1)
	expectMaxBatch(mock, 1)
	rows := sqlmock.NewRows([]string{"migration", "batch", "created_at"}).
		AddRow("20240101000000_create_users", 1, time.Now()).
		AddRow("20240102000000_create_posts", 1, time.Now())
	mock.ExpectQuery("SELECT migration, batch, created_at FROM .* WHERE batch").
		WithArgs(1).WillReturnRows(rows)

	// Rollback in reverse order: posts first, then users
	expectMigrationTx(mock)
	expectRemove(mock, "20240102000000_create_posts")

	expectMigrationTx(mock)
	expectRemove(mock, "20240101000000_create_users")

	err := m.Rollback(0)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRollback_BySteps(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	require.NoError(t, m.Register("20240101000000_create_users", &noopMigration{}))
	require.NoError(t, m.Register("20240102000000_create_posts", &noopMigration{}))
	require.NoError(t, m.Register("20240103000000_create_tags", &noopMigration{}))

	expectEnsureTable(mock)

	// GetLastNMigrations(2): GetApplied returns all 3, we take last 2 reversed
	expectGetApplied(mock, []MigrationRecord{
		{Name: "20240101000000_create_users", Batch: 1, CreatedAt: time.Now()},
		{Name: "20240102000000_create_posts", Batch: 1, CreatedAt: time.Now()},
		{Name: "20240103000000_create_tags", Batch: 2, CreatedAt: time.Now()},
	})

	// Rollback: tags first, then posts (reverse order)
	expectMigrationTx(mock)
	expectRemove(mock, "20240103000000_create_tags")

	expectMigrationTx(mock)
	expectRemove(mock, "20240102000000_create_posts")

	err := m.Rollback(2)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Reset Tests ---

func TestReset(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	require.NoError(t, m.Register("20240101000000_create_users", &noopMigration{}))
	require.NoError(t, m.Register("20240102000000_create_posts", &noopMigration{}))

	expectEnsureTable(mock)
	expectGetApplied(mock, []MigrationRecord{
		{Name: "20240101000000_create_users", Batch: 1, CreatedAt: time.Now()},
		{Name: "20240102000000_create_posts", Batch: 1, CreatedAt: time.Now()},
	})

	// Reverse order: posts first, then users
	expectMigrationTx(mock)
	expectRemove(mock, "20240102000000_create_posts")

	expectMigrationTx(mock)
	expectRemove(mock, "20240101000000_create_users")

	err := m.Reset()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Status Tests ---

func TestStatus(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	now := time.Now()
	require.NoError(t, m.Register("20240101000000_create_users", &noopMigration{}))
	require.NoError(t, m.Register("20240102000000_create_posts", &noopMigration{}))

	expectEnsureTable(mock)
	expectGetApplied(mock, []MigrationRecord{
		{Name: "20240101000000_create_users", Batch: 1, CreatedAt: now},
	})

	statuses, err := m.Status()
	require.NoError(t, err)
	require.Len(t, statuses, 2)

	assert.Equal(t, "20240101000000_create_users", statuses[0].Name)
	assert.True(t, statuses[0].Applied)
	assert.Equal(t, 1, statuses[0].Batch)
	assert.NotNil(t, statuses[0].AppliedAt)

	assert.Equal(t, "20240102000000_create_posts", statuses[1].Name)
	assert.False(t, statuses[1].Applied)
	assert.Equal(t, 0, statuses[1].Batch)
	assert.Nil(t, statuses[1].AppliedAt)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Fresh Tests ---

func TestFresh(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	require.NoError(t, m.Register("20240101000000_create_users", &noopMigration{}))

	// Drop all tables
	mock.ExpectExec("DROP ALL").WillReturnResult(sqlmock.NewResult(0, 0))

	// Then Up()
	expectEnsureTable(mock)
	expectGetApplied(mock, nil)
	expectMaxBatch(mock, 0)
	expectMigrationTx(mock)
	expectRecord(mock, "20240101000000_create_users", 1)

	err := m.Fresh()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFresh_DropFails(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	mock.ExpectExec("DROP ALL").WillReturnError(errors.New("permission denied"))

	err := m.Fresh()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "drop all tables")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFresh_NoGrammar(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	m := New(db) // no grammar
	err = m.Fresh()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fresh requires a grammar")
}

// --- Refresh Tests ---

func TestRefresh(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	require.NoError(t, m.Register("20240101000000_create_users", &noopMigration{}))

	// Reset phase: EnsureTable + GetApplied (1 applied) + Down + Remove
	expectEnsureTable(mock)
	expectGetApplied(mock, []MigrationRecord{
		{Name: "20240101000000_create_users", Batch: 1, CreatedAt: time.Now()},
	})
	expectMigrationTx(mock)
	expectRemove(mock, "20240101000000_create_users")

	// Up phase: EnsureTable + GetApplied (empty) + NextBatch + Execute + Record
	expectEnsureTable(mock)
	expectGetApplied(mock, nil)
	expectMaxBatch(mock, 0)
	expectMigrationTx(mock)
	expectRecord(mock, "20240101000000_create_users", 1)

	err := m.Refresh()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Hook Integration Tests ---

func TestUp_BeforeHookError_AbortsMigration(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	require.NoError(t, m.Register("20240101000000_create_users", &noopMigration{}))
	m.BeforeMigrate(func(name, direction string) error {
		return errors.New("hook failed")
	})

	expectEnsureTable(mock)
	expectGetApplied(mock, nil)
	expectMaxBatch(mock, 0)

	err := m.Up()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hook failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUp_AfterHookCalled(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	var hookCalled bool
	require.NoError(t, m.Register("20240101000000_create_users", &noopMigration{}))
	m.AfterMigrate(func(name, direction string, d time.Duration) error {
		hookCalled = true
		assert.Equal(t, "20240101000000_create_users", name)
		assert.Equal(t, "up", direction)
		return nil
	})

	expectEnsureTable(mock)
	expectGetApplied(mock, nil)
	expectMaxBatch(mock, 0)
	expectMigrationTx(mock)
	expectRecord(mock, "20240101000000_create_users", 1)

	err := m.Up()
	assert.NoError(t, err)
	assert.True(t, hookCalled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Option Tests ---

func TestWithTableName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	m := New(db, WithGrammar(&mockGrammar{}), WithTableName("custom_migrations"))
	require.NoError(t, m.Register("20240101000000_create_users", &noopMigration{}))

	// EnsureTable should use custom table name
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS custom_migrations").
		WillReturnResult(sqlmock.NewResult(0, 0))

	// GetApplied from custom table
	mock.ExpectQuery("SELECT migration, batch, created_at FROM custom_migrations").
		WillReturnRows(sqlmock.NewRows([]string{"migration", "batch", "created_at"}))

	// NextBatchNumber
	mock.ExpectQuery("SELECT COALESCE").WillReturnRows(
		sqlmock.NewRows([]string{"max"}).AddRow(0),
	)

	expectMigrationTx(mock)

	// Record into custom table
	mock.ExpectExec("INSERT INTO custom_migrations").
		WithArgs("20240101000000_create_users", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = m.Up()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Rollback edge case: no applied migrations ---

func TestRollback_NoApplied(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	expectEnsureTable(mock)
	expectMaxBatch(mock, 0) // no batches

	err := m.Rollback(0)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReset_NoApplied(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	expectEnsureTable(mock)
	expectGetApplied(mock, nil)

	err := m.Reset()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Edge-case tests (Task 8.4) ---

// failingUpMigration succeeds on Down but fails on Up.
type failingUpMigration struct {
	downCalled bool
}

func (m *failingUpMigration) Up(s *schema.Builder) error   { return errors.New("up failed") }
func (m *failingUpMigration) Down(s *schema.Builder) error { m.downCalled = true; return nil }

// failingDownMigration succeeds on Up but fails on Down.
type failingDownMigration struct {
	upCalled bool
}

func (m *failingDownMigration) Up(s *schema.Builder) error   { m.upCalled = true; return nil }
func (m *failingDownMigration) Down(s *schema.Builder) error { return errors.New("down failed") }

// TestUp_FailureStopsSubsequent verifies that when a migration's Up() fails,
// subsequent migrations are not executed and only prior successes are recorded.
// Validates: Requirement 2.4
func TestUp_FailureStopsSubsequent(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	first := &noopMigration{}
	second := &failingUpMigration{}
	third := &noopMigration{}

	require.NoError(t, m.Register("20240101000000_first", first))
	require.NoError(t, m.Register("20240102000000_second", second))
	require.NoError(t, m.Register("20240103000000_third", third))

	expectEnsureTable(mock)
	expectGetApplied(mock, nil)
	expectMaxBatch(mock, 0)

	// First migration succeeds
	expectMigrationTx(mock)
	expectRecord(mock, "20240101000000_first", 1)

	// Second migration fails: begin + rollback (no commit, no record)
	mock.ExpectBegin()
	mock.ExpectRollback()

	err := m.Up()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "20240102000000_second")

	// First ran, third never ran
	assert.True(t, first.upCalled)
	assert.False(t, third.upCalled)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRollback_FailureStopsFurtherRollback verifies that when a migration's
// Down() fails during rollback, earlier migrations are not rolled back.
// Validates: Requirement 3.5
func TestRollback_FailureStopsFurtherRollback(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	first := &noopMigration{}
	second := &noopMigration{}
	third := &failingDownMigration{}

	require.NoError(t, m.Register("20240101000000_first", first))
	require.NoError(t, m.Register("20240102000000_second", second))
	require.NoError(t, m.Register("20240103000000_third", third))

	expectEnsureTable(mock)

	// Rollback by batch: all 3 in batch 1
	expectMaxBatch(mock, 1)
	rows := sqlmock.NewRows([]string{"migration", "batch", "created_at"}).
		AddRow("20240101000000_first", 1, time.Now()).
		AddRow("20240102000000_second", 1, time.Now()).
		AddRow("20240103000000_third", 1, time.Now())
	mock.ExpectQuery("SELECT migration, batch, created_at FROM .* WHERE batch").
		WithArgs(1).WillReturnRows(rows)

	// Reverse order: third rolls back first — but it fails
	mock.ExpectBegin()
	mock.ExpectRollback()

	err := m.Rollback(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "20240103000000_third")

	// Second and first should NOT have been rolled back
	assert.False(t, second.downCalled)
	assert.False(t, first.downCalled)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRefresh_UpPhaseFailure verifies that when the up phase of Refresh fails,
// the error message contains the name of the failing migration.
// Validates: Requirement 4.5
func TestRefresh_UpPhaseFailure(t *testing.T) {
	m, db, mock := newTestMigrator(t)
	defer db.Close()

	require.NoError(t, m.Register("20240101000000_first", &noopMigration{}))
	require.NoError(t, m.Register("20240102000000_second", &failingUpMigration{}))

	// Reset phase: first migration is applied, roll it back
	expectEnsureTable(mock)
	expectGetApplied(mock, []MigrationRecord{
		{Name: "20240101000000_first", Batch: 1, CreatedAt: time.Now()},
	})
	expectMigrationTx(mock)
	expectRemove(mock, "20240101000000_first")

	// Up phase: both pending, first succeeds, second fails
	expectEnsureTable(mock)
	expectGetApplied(mock, nil)
	expectMaxBatch(mock, 0)

	// First migration succeeds
	expectMigrationTx(mock)
	expectRecord(mock, "20240101000000_first", 1)

	// Second migration fails
	mock.ExpectBegin()
	mock.ExpectRollback()

	err := m.Refresh()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh up phase")
	assert.Contains(t, err.Error(), "20240102000000_second")

	assert.NoError(t, mock.ExpectationsWereMet())
}
