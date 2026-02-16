package database

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// Req 10.1: Begin a database transaction before calling the migration method
// Req 10.2: Commit the transaction when the migration method returns nil
func TestWithTransaction_Success(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectCommit()

	called := false
	err := WithTransaction(db, func(tx *sql.Tx) error {
		called = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, called, "fn should have been called")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Req 10.3: Roll back the transaction when the migration method returns an error
func TestWithTransaction_FnError_Rollback(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectRollback()

	fnErr := errors.New("migration failed")
	err := WithTransaction(db, func(tx *sql.Tx) error {
		return fnErr
	})

	assert.ErrorIs(t, err, fnErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Req 10.5: If the transaction commit fails, attempt a rollback and return a descriptive error.
// Note: In Go's database/sql, a failed Commit marks the tx as done, so the
// subsequent Rollback returns sql.ErrTxDone (never reaches the driver).
// The implementation still attempts rollback as required by the spec.
func TestWithTransaction_CommitFails_ReturnsDescriptiveError(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit error"))

	err := WithTransaction(db, func(tx *sql.Tx) error {
		return nil
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTransactionFailed)
	assert.Contains(t, err.Error(), "commit")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTransaction_BeginFails(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectBegin().WillReturnError(fmt.Errorf("begin error"))

	err := WithTransaction(db, func(tx *sql.Tx) error {
		t.Fatal("fn should not be called when begin fails")
		return nil
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTransactionFailed)
	assert.Contains(t, err.Error(), "begin")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTransaction_FnError_RollbackFails(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectRollback().WillReturnError(fmt.Errorf("rollback error"))

	fnErr := errors.New("migration failed")
	err := WithTransaction(db, func(tx *sql.Tx) error {
		return fnErr
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTransactionFailed)
	assert.Contains(t, err.Error(), "migration failed")
	assert.Contains(t, err.Error(), "rollback error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Verify the fn receives a valid *sql.Tx that can execute queries
func TestWithTransaction_FnReceivesTx(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO test").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := WithTransaction(db, func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO test VALUES (1)")
		return err
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
