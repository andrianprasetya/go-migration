package migrator

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test AutoDiscover() on empty auto-registry returns nil.
// Validates: Requirement 5.2
func TestAutoDiscover_EmptyAutoRegistry_ReturnsNil(t *testing.T) {
	defer ResetAutoRegistry()

	m := New(nil)
	err := m.AutoDiscover()

	assert.NoError(t, err, "AutoDiscover on empty auto-registry should return nil")
	assert.Empty(t, m.registry.GetAll(), "internal registry should remain empty")
}

// Test WithAutoDiscover() with empty auto-registry creates Migrator without error.
// Validates: Requirement 6.3
func TestWithAutoDiscover_EmptyAutoRegistry_CreatesWithoutError(t *testing.T) {
	defer ResetAutoRegistry()

	assert.NotPanics(t, func() {
		m := New(nil, WithAutoDiscover())
		assert.NotNil(t, m, "Migrator should be created successfully")
		assert.Empty(t, m.registry.GetAll(), "internal registry should be empty")
	})
}

// Test backward compatibility: Register() still works after AutoDiscover().
// Validates: Requirement 2.5
func TestRegister_StillWorksAfterAutoDiscover(t *testing.T) {
	defer ResetAutoRegistry()

	AutoRegister("20240101000001_create_users", &stubMigration{})

	m := New(nil)
	err := m.AutoDiscover()
	require.NoError(t, err)

	// Register() should still work after AutoDiscover()
	err = m.Register("20240101000002_create_posts", &stubMigration{})
	assert.NoError(t, err, "Register should still work after AutoDiscover")

	all := m.registry.GetAll()
	assert.Len(t, all, 2)
	assert.Equal(t, "20240101000001_create_users", all[0].Name)
	assert.Equal(t, "20240101000002_create_posts", all[1].Name)
}

// Test AutoDiscover() and Register() used together with no overlap.
// Validates: Requirement 4.1
func TestAutoDiscover_AndRegister_NoOverlap(t *testing.T) {
	defer ResetAutoRegistry()

	// Auto-register two migrations
	AutoRegister("20240101000001_create_users", &stubMigration{})
	AutoRegister("20240101000003_create_comments", &stubMigration{})

	m := New(nil)

	// Manually register a different migration
	err := m.Register("20240101000002_create_posts", &stubMigration{})
	require.NoError(t, err)

	// AutoDiscover should succeed with no overlap
	err = m.AutoDiscover()
	assert.NoError(t, err, "AutoDiscover should succeed when no name overlap exists")

	// All three migrations should be present in sorted order
	all := m.registry.GetAll()
	assert.Len(t, all, 3)
	assert.Equal(t, "20240101000001_create_users", all[0].Name)
	assert.Equal(t, "20240101000002_create_posts", all[1].Name)
	assert.Equal(t, "20240101000003_create_comments", all[2].Name)
}

// Test AutoDiscover() and Register() with overlapping name returns error.
// Validates: Requirement 4.3
func TestAutoDiscover_AndRegister_OverlappingName_ReturnsError(t *testing.T) {
	defer ResetAutoRegistry()

	conflictName := "20240101000001_create_users"

	// Auto-register a migration
	AutoRegister(conflictName, &stubMigration{})

	m := New(nil)

	// Manually register the same name
	err := m.Register(conflictName, &stubMigration{})
	require.NoError(t, err)

	// AutoDiscover should return error wrapping ErrDuplicateMigration
	err = m.AutoDiscover()
	assert.Error(t, err, "AutoDiscover should return error for overlapping name")
	assert.True(t, errors.Is(err, ErrDuplicateMigration),
		"error should wrap ErrDuplicateMigration, got: %v", err)
	assert.Contains(t, err.Error(), conflictName,
		"error message should contain the conflicting name")
}
