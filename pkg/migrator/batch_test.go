package migrator

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBatchManager(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	tracker := NewTracker(db, "migrations")
	bm := NewBatchManager(tracker)
	assert.NotNil(t, bm)
}

// --- NextBatchNumber ---

func TestNextBatchNumber_FirstBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(batch\\), 0\\) FROM migrations").
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(0))

	bm := NewBatchManager(NewTracker(db, "migrations"))
	next, err := bm.NextBatchNumber()
	require.NoError(t, err)
	assert.Equal(t, 1, next)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNextBatchNumber_SubsequentBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(batch\\), 0\\) FROM migrations").
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(5))

	bm := NewBatchManager(NewTracker(db, "migrations"))
	next, err := bm.NextBatchNumber()
	require.NoError(t, err)
	assert.Equal(t, 6, next)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNextBatchNumber_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(batch\\), 0\\) FROM migrations").
		WillReturnError(errors.New("db error"))

	bm := NewBatchManager(NewTracker(db, "migrations"))
	_, err = bm.NextBatchNumber()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTrackingTable))
}

// --- GetLastBatch ---

func TestGetLastBatch_WithRecords(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Now().Truncate(time.Second)

	// GetLastBatchNumber
	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(batch\\), 0\\) FROM migrations").
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(3))

	// GetByBatch(3)
	mock.ExpectQuery("SELECT migration, batch, created_at FROM migrations WHERE batch = \\$1").
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"migration", "batch", "created_at"}).
			AddRow("20240315000000_add_posts", 3, now).
			AddRow("20240316000000_add_comments", 3, now))

	bm := NewBatchManager(NewTracker(db, "migrations"))
	records, err := bm.GetLastBatch()
	require.NoError(t, err)
	require.Len(t, records, 2)
	assert.Equal(t, "20240315000000_add_posts", records[0].Name)
	assert.Equal(t, "20240316000000_add_comments", records[1].Name)
	assert.Equal(t, 3, records[0].Batch)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLastBatch_NoRecords(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(batch\\), 0\\) FROM migrations").
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(0))

	bm := NewBatchManager(NewTracker(db, "migrations"))
	records, err := bm.GetLastBatch()
	require.NoError(t, err)
	assert.Nil(t, records)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLastBatch_BatchNumberError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(batch\\), 0\\) FROM migrations").
		WillReturnError(errors.New("db error"))

	bm := NewBatchManager(NewTracker(db, "migrations"))
	_, err = bm.GetLastBatch()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTrackingTable))
}

func TestGetLastBatch_GetByBatchError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT COALESCE\\(MAX\\(batch\\), 0\\) FROM migrations").
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(2))

	mock.ExpectQuery("SELECT migration, batch, created_at FROM migrations WHERE batch = \\$1").
		WithArgs(2).
		WillReturnError(errors.New("db error"))

	bm := NewBatchManager(NewTracker(db, "migrations"))
	_, err = bm.GetLastBatch()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTrackingTable))
}

// --- GetLastNMigrations ---

func TestGetLastNMigrations_ReturnsReversed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Now().Truncate(time.Second)
	mock.ExpectQuery("SELECT migration, batch, created_at FROM migrations ORDER BY migration ASC").
		WillReturnRows(sqlmock.NewRows([]string{"migration", "batch", "created_at"}).
			AddRow("20240101000000_first", 1, now).
			AddRow("20240201000000_second", 1, now).
			AddRow("20240301000000_third", 2, now).
			AddRow("20240401000000_fourth", 2, now))

	bm := NewBatchManager(NewTracker(db, "migrations"))
	records, err := bm.GetLastNMigrations(2)
	require.NoError(t, err)
	require.Len(t, records, 2)
	// Should be in reverse timestamp order
	assert.Equal(t, "20240401000000_fourth", records[0].Name)
	assert.Equal(t, "20240301000000_third", records[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLastNMigrations_NExceedsTotal(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Now().Truncate(time.Second)
	mock.ExpectQuery("SELECT migration, batch, created_at FROM migrations ORDER BY migration ASC").
		WillReturnRows(sqlmock.NewRows([]string{"migration", "batch", "created_at"}).
			AddRow("20240101000000_first", 1, now).
			AddRow("20240201000000_second", 1, now))

	bm := NewBatchManager(NewTracker(db, "migrations"))
	records, err := bm.GetLastNMigrations(10)
	require.NoError(t, err)
	require.Len(t, records, 2)
	assert.Equal(t, "20240201000000_second", records[0].Name)
	assert.Equal(t, "20240101000000_first", records[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLastNMigrations_ZeroN(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	bm := NewBatchManager(NewTracker(db, "migrations"))
	records, err := bm.GetLastNMigrations(0)
	require.NoError(t, err)
	assert.Nil(t, records)
}

func TestGetLastNMigrations_NegativeN(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	bm := NewBatchManager(NewTracker(db, "migrations"))
	records, err := bm.GetLastNMigrations(-1)
	require.NoError(t, err)
	assert.Nil(t, records)
}

func TestGetLastNMigrations_NoApplied(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT migration, batch, created_at FROM migrations ORDER BY migration ASC").
		WillReturnRows(sqlmock.NewRows([]string{"migration", "batch", "created_at"}))

	bm := NewBatchManager(NewTracker(db, "migrations"))
	records, err := bm.GetLastNMigrations(5)
	require.NoError(t, err)
	assert.Nil(t, records)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetLastNMigrations_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT migration, batch, created_at FROM migrations ORDER BY migration ASC").
		WillReturnError(errors.New("db error"))

	bm := NewBatchManager(NewTracker(db, "migrations"))
	_, err = bm.GetLastNMigrations(3)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTrackingTable))
}
