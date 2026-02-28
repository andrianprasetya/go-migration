package migrator

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/andrianprasetya/go-migration/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test helpers ---

// mockGrammar implements schema.Grammar for testing purposes.
type mockGrammar struct{}

func (g *mockGrammar) CompileCreate(bp *schema.Blueprint) (string, error) {
	return fmt.Sprintf("CREATE TABLE %s ()", bp.Table()), nil
}
func (g *mockGrammar) CompileAlter(bp *schema.Blueprint) ([]string, error) {
	return []string{fmt.Sprintf("ALTER TABLE %s", bp.Table())}, nil
}
func (g *mockGrammar) CompileDrop(table string) string { return "DROP TABLE " + table }
func (g *mockGrammar) CompileDropIfExists(table string) string {
	return "DROP TABLE IF EXISTS " + table
}
func (g *mockGrammar) CompileRename(from, to string) string {
	return "ALTER TABLE " + from + " RENAME TO " + to
}
func (g *mockGrammar) CompileHasTable(table string) string       { return "SELECT 1" }
func (g *mockGrammar) CompileHasColumn(table, col string) string { return "SELECT 1" }
func (g *mockGrammar) CompileDropAllTables() string              { return "DROP ALL" }
func (g *mockGrammar) CompileColumnType(col schema.ColumnDefinition) (string, error) {
	return "TEXT", nil
}

// noopMigration succeeds on both Up and Down.
type noopMigration struct {
	upCalled   bool
	downCalled bool
}

func (m *noopMigration) Up(s *schema.Builder) error   { m.upCalled = true; return nil }
func (m *noopMigration) Down(s *schema.Builder) error { m.downCalled = true; return nil }

// failingMigration returns an error on both Up and Down.
type failingMigration struct{}

func (m *failingMigration) Up(s *schema.Builder) error   { return errors.New("up failed") }
func (m *failingMigration) Down(s *schema.Builder) error { return errors.New("down failed") }

// noTxMigration opts out of transaction wrapping.
type noTxMigration struct {
	upCalled   bool
	downCalled bool
}

func (m *noTxMigration) Up(s *schema.Builder) error   { m.upCalled = true; return nil }
func (m *noTxMigration) Down(s *schema.Builder) error { m.downCalled = true; return nil }
func (m *noTxMigration) DisableTransaction() bool     { return true }

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return db, mock
}

// --- Tests ---

func TestExecuteInTransaction_Up(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	runner := NewRunner(db, &mockGrammar{}, nil)
	m := &noopMigration{}

	err := runner.Execute(m, "up")

	assert.NoError(t, err)
	assert.True(t, m.upCalled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteInTransaction_Down(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	runner := NewRunner(db, &mockGrammar{}, nil)
	m := &noopMigration{}

	err := runner.Execute(m, "down")

	assert.NoError(t, err)
	assert.True(t, m.downCalled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteWithoutTransaction_OptOut(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	// No Begin/Commit expected — migration runs directly on db
	runner := NewRunner(db, &mockGrammar{}, nil)
	m := &noTxMigration{}

	err := runner.Execute(m, "up")

	assert.NoError(t, err)
	assert.True(t, m.upCalled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteInTransaction_MigrationError_TriggersRollback(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	runner := NewRunner(db, &mockGrammar{}, nil)
	m := &failingMigration{}

	err := runner.Execute(m, "up")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "up failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteInTransaction_CommitFailure_ReturnsTransactionFailed(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	runner := NewRunner(db, &mockGrammar{}, nil)
	m := &noopMigration{}

	err := runner.Execute(m, "up")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTransactionFailed))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteInTransaction_UnknownDirection(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	runner := NewRunner(db, &mockGrammar{}, nil)
	m := &noopMigration{}

	err := runner.Execute(m, "sideways")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown migration direction")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteInTransaction_BeginError(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectBegin().WillReturnError(errors.New("begin error"))

	runner := NewRunner(db, &mockGrammar{}, nil)
	m := &noopMigration{}

	err := runner.ExecuteInTransaction(m, "up")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Dry-run tests ---

// createTableMigration creates a table in Up and drops it in Down.
type createTableMigration struct{}

func (m *createTableMigration) Up(b *schema.Builder) error {
	return b.Create("users", func(bp *schema.Blueprint) {
		bp.String("name", 255)
	})
}

func (m *createTableMigration) Down(b *schema.Builder) error {
	return b.Drop("users")
}

func TestExecuteDryRun_WritesSQL_NoTransaction(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	var buf bytes.Buffer
	runner := NewRunner(db, &mockGrammar{}, nil)
	runner.SetDryRun(&buf)

	m := &noopMigration{}
	err := runner.Execute(m, "up")

	assert.NoError(t, err)
	assert.True(t, m.upCalled)
	// No Begin/Commit should have been called on the mock DB.
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteDryRun_WritesMigrationNamePrefix(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	var buf bytes.Buffer
	runner := NewRunner(db, &mockGrammar{}, nil)
	runner.SetDryRun(&buf)

	m := &createTableMigration{}
	err := runner.ExecuteDryRun(m, "up", "20240101_create_users")

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "-- Migration: 20240101_create_users")
	assert.Contains(t, output, "CREATE TABLE")
	// No DB interactions expected.
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteDryRun_Down_WritesDropSQL(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	var buf bytes.Buffer
	runner := NewRunner(db, &mockGrammar{}, nil)
	runner.SetDryRun(&buf)

	m := &createTableMigration{}
	err := runner.ExecuteDryRun(m, "down", "20240101_create_users")

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "-- Migration: 20240101_create_users")
	assert.Contains(t, output, "DROP TABLE")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteDryRun_NoTransactionCommitted(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	var buf bytes.Buffer
	runner := NewRunner(db, &mockGrammar{}, nil)
	runner.SetDryRun(&buf)

	// Even a migration that would normally use transactions should not
	// start any transaction in dry-run mode.
	m := &noopMigration{}
	err := runner.Execute(m, "up")

	assert.NoError(t, err)
	// mock has no expectations set — if Begin/Commit were called, it would fail.
	assert.NoError(t, mock.ExpectationsWereMet())
}
