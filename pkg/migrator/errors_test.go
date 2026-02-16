package migrator

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSentinelErrors_ImplementErrorInterface(t *testing.T) {
	sentinels := []error{
		ErrConnectionFailed,
		ErrMigrationNotFound,
		ErrDuplicateMigration,
		ErrInvalidMigrationName,
		ErrTransactionFailed,
		ErrTrackingTable,
		ErrDuplicateSeeder,
		ErrInvalidSeederName,
		ErrCircularDependency,
		ErrSeederNotFound,
		ErrUnsupportedType,
		ErrConnectionNotFound,
		ErrConfigValidation,
	}

	for _, sentinel := range sentinels {
		assert.NotEmpty(t, sentinel.Error(), "sentinel error should have a non-empty message")
	}
}

func TestSentinelErrors_UniqueMessages(t *testing.T) {
	sentinels := []error{
		ErrConnectionFailed,
		ErrMigrationNotFound,
		ErrDuplicateMigration,
		ErrInvalidMigrationName,
		ErrTransactionFailed,
		ErrTrackingTable,
		ErrDuplicateSeeder,
		ErrInvalidSeederName,
		ErrCircularDependency,
		ErrSeederNotFound,
		ErrUnsupportedType,
		ErrConnectionNotFound,
		ErrConfigValidation,
	}

	seen := make(map[string]bool)
	for _, sentinel := range sentinels {
		msg := sentinel.Error()
		assert.False(t, seen[msg], "duplicate error message: %s", msg)
		seen[msg] = true
	}
}

func TestSentinelErrors_ErrorsIs(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
	}{
		{"ErrConnectionFailed", ErrConnectionFailed},
		{"ErrMigrationNotFound", ErrMigrationNotFound},
		{"ErrDuplicateMigration", ErrDuplicateMigration},
		{"ErrInvalidMigrationName", ErrInvalidMigrationName},
		{"ErrTransactionFailed", ErrTransactionFailed},
		{"ErrTrackingTable", ErrTrackingTable},
		{"ErrDuplicateSeeder", ErrDuplicateSeeder},
		{"ErrInvalidSeederName", ErrInvalidSeederName},
		{"ErrCircularDependency", ErrCircularDependency},
		{"ErrSeederNotFound", ErrSeederNotFound},
		{"ErrUnsupportedType", ErrUnsupportedType},
		{"ErrConnectionNotFound", ErrConnectionNotFound},
		{"ErrConfigValidation", ErrConfigValidation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Direct match
			assert.True(t, errors.Is(tt.sentinel, tt.sentinel))

			// Wrapped error preserves chain
			wrapped := fmt.Errorf("operation failed: %w", tt.sentinel)
			assert.True(t, errors.Is(wrapped, tt.sentinel),
				"wrapped error should match sentinel via errors.Is()")
			assert.Contains(t, wrapped.Error(), tt.sentinel.Error(),
				"wrapped error should contain original message")
		})
	}
}

func TestSentinelErrors_WrappedPreservesContext(t *testing.T) {
	wrapped := fmt.Errorf("migration %q up: %w", "20240101_create_users", ErrTransactionFailed)

	assert.True(t, errors.Is(wrapped, ErrTransactionFailed))
	assert.Contains(t, wrapped.Error(), "20240101_create_users")
	assert.Contains(t, wrapped.Error(), "transaction failed")
}
