package migrator

import (
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
