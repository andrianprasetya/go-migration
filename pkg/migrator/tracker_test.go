package migrator

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTracker(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	tracker := NewTracker(db, "migrations")
	assert.NotNil(t, tracker)
}

func TestEnsureTable_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS migrations").
		WillReturnResult(sqlmock.NewResult(0, 0))

	tracker := NewTracker(db, "migrations")
	err = tracker.EnsureTable()
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEnsureTable_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS migrations").
		WillReturnError(errors.New("db error"))

	tracker := NewTracker(db, "migrations")
	err = tracker.EnsureTable()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTrackingTable))
}

func TestEnsureTable_CustomTableName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS custom_migrations").
		WillReturnResult(sqlmock.NewResult(0, 0))

	tracker := NewTracker(db, "custom_migrations")
	err = tracker.EnsureTable()
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetApplied_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Now().Truncate(time.Second)
	rows := sqlmock.NewRows([]string{"migration", "batch", "created_at"}).
		AddRow("20240115000000_first", 1, now).
		AddRow("20240215000000_second", 1, now)

	mock.ExpectQuery("SELECT migration, batch, created_at FROM migrations ORDER BY migration ASC").
		WillReturnRows(rows)

	tracker := NewTracker(db, "migrations")
	records, err := tracker.GetApplied()
	require.NoError(t, err)
	require.Len(t, records, 2)
	assert.Equal(t, "20240115000000_first", records[0].Name)
	assert.Equal(t, 1, records[0].Batch)
	assert.Equal(t, "20240215000000_second", records[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetApplied_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"migration", "batch", "created_at"})
	mock.ExpectQuery("SELECT migration, batch, created_at FROM migrations ORDER BY migration ASC").
		WillReturnRows(rows)

	tracker := NewTracker(db, "migrations")
	records, err := tracker.GetApplied()
	require.NoError(t, err)
	assert.Empty(t, records)
}

func TestGetApplied_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT migration, batch, created_at FROM migrations").
		WillReturnError(errors.New("db error"))

	tracker := NewTracker(db, "migrations")
	_, err = tracker.GetApplied()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTrackingTable))
}

func TestGetByBatch_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Now().Truncate(time.Second)
	rows := sqlmock.NewRows([]string{"migration", "batch", "created_at"}).
		AddRow("20240115000000_first", 2, now).
		AddRow("20240215000000_second", 2, now)

	mock.ExpectQuery("SELECT migration, batch, created_at FROM migrations WHERE batch = \\$1 ORDER BY migration ASC").
		WithArgs(2).
		WillReturnRows(rows)

	tracker := NewTracker(db, "migrations")
	records, err := tracker.GetByBatch(2)
	require.NoError(t, err)
	require.Len(t, records, 2)
	assert.Equal(t, 2, records[0].Batch)
	assert.Equal(t, 2, records[1].Batch)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByBatch_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"migration", "batch", "created_at"})
	mock.ExpectQuery("SELECT migration, batch, created_at FROM migrations WHERE batch = \\$1").
		WithArgs(99).
		WillReturnRows(rows)

	tracker := NewTracker(db, "migrations")
	records, err := tracker.GetByBatch(99)
	require.NoError(t, err)
	assert.Empty(t, records)
}

func TestGetByBatch_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT migration, batch, created_at FROM migrations WHERE batch = \\$1").
		WithArgs(1).
		WillReturnError(errors.New("db error"))

	tracker := NewTracker(db, "migrations")
	_, err = tracker.GetByBatch(1)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTrackingTable))
}

func TestGetLastBatchNumber_WithRecords(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"max"}).AddRow(3)
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(batch\\), 0\\) FROM migrations").
		WillReturnRows(rows)

	tracker := NewTracker(db, "migrations")
	batch, err := tracker.GetLastBatchNumber()
	require.NoError(t, err)
	assert.Equal(t, 3, batch)
}

func TestGetLastBatchNumber_NoRecords(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"max"}).AddRow(0)
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(batch\\), 0\\) FROM migrations").
		WillReturnRows(rows)

	tracker := NewTracker(db, "migrations")
	batch, err := tracker.GetLastBatchNumber()
	require.NoError(t, err)
	assert.Equal(t, 0, batch)
}

func TestGetLastBatchNumber_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(batch\\), 0\\) FROM migrations").
		WillReturnError(errors.New("db error"))

	tracker := NewTracker(db, "migrations")
	_, err = tracker.GetLastBatchNumber()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTrackingTable))
}

func TestRecord_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("INSERT INTO migrations").
		WithArgs("20240115000000_create_users", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	tracker := NewTracker(db, "migrations")
	err = tracker.Record("20240115000000_create_users", 1)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRecord_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("INSERT INTO migrations").
		WithArgs("20240115000000_create_users", 1).
		WillReturnError(errors.New("duplicate key"))

	tracker := NewTracker(db, "migrations")
	err = tracker.Record("20240115000000_create_users", 1)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTrackingTable))
	assert.Contains(t, err.Error(), "20240115000000_create_users")
}

func TestRemove_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("DELETE FROM migrations WHERE migration = \\$1").
		WithArgs("20240115000000_create_users").
		WillReturnResult(sqlmock.NewResult(0, 1))

	tracker := NewTracker(db, "migrations")
	err = tracker.Remove("20240115000000_create_users")
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRemove_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("DELETE FROM migrations WHERE migration = \\$1").
		WithArgs("20240115000000_create_users").
		WillReturnError(errors.New("db error"))

	tracker := NewTracker(db, "migrations")
	err = tracker.Remove("20240115000000_create_users")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTrackingTable))
	assert.Contains(t, err.Error(), "20240115000000_create_users")
}
