package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to create empty files in a temp directory.
func touch(t *testing.T, dir, name string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), nil, 0o644))
}

func TestScanMigrations(t *testing.T) {
	dir := t.TempDir()

	// Valid migration files
	touch(t, dir, "20240101120000_create_users.go")
	touch(t, dir, "20240201120000_add_email.go")
	// Not a migration (no timestamp prefix)
	touch(t, dir, "helper.go")
	// Test file should be excluded
	touch(t, dir, "20240101120000_create_users_test.go")
	// Subdirectory should be ignored
	require.NoError(t, os.Mkdir(filepath.Join(dir, "20240301120000_subdir.go"), 0o755))

	files, err := ScanMigrations(dir)
	require.NoError(t, err)

	expected := []string{
		filepath.Join(dir, "20240101120000_create_users.go"),
		filepath.Join(dir, "20240201120000_add_email.go"),
	}
	assert.Equal(t, expected, files)
}

func TestScanMigrations_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	files, err := ScanMigrations(dir)
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestScanMigrations_NonExistentDir(t *testing.T) {
	_, err := ScanMigrations("/nonexistent_dir_12345")
	assert.Error(t, err)
}

func TestScanSeeders(t *testing.T) {
	dir := t.TempDir()

	// Valid seeder files
	touch(t, dir, "users_seeder.go")
	touch(t, dir, "posts_seeder.go")
	// Not a seeder
	touch(t, dir, "helper.go")
	// Test file should be excluded
	touch(t, dir, "users_seeder_test.go")

	files, err := ScanSeeders(dir)
	require.NoError(t, err)

	expected := []string{
		filepath.Join(dir, "posts_seeder.go"),
		filepath.Join(dir, "users_seeder.go"),
	}
	assert.Equal(t, expected, files)
}

func TestScanSeeders_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	files, err := ScanSeeders(dir)
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestScanSeeders_NonExistentDir(t *testing.T) {
	_, err := ScanSeeders("/nonexistent_dir_12345")
	assert.Error(t, err)
}

func TestScanMigrations_SortOrder(t *testing.T) {
	dir := t.TempDir()

	// Create files in reverse order
	touch(t, dir, "20240301120000_third.go")
	touch(t, dir, "20240101120000_first.go")
	touch(t, dir, "20240201120000_second.go")

	files, err := ScanMigrations(dir)
	require.NoError(t, err)

	expected := []string{
		filepath.Join(dir, "20240101120000_first.go"),
		filepath.Join(dir, "20240201120000_second.go"),
		filepath.Join(dir, "20240301120000_third.go"),
	}
	assert.Equal(t, expected, files)
}
