package migrator

import (
	"errors"
	"testing"

	"github.com/andrianprasetya/go-migration/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubMigration is a no-op Migration for testing.
type stubMigration struct{}

func (s *stubMigration) Up(_ *schema.Builder) error   { return nil }
func (s *stubMigration) Down(_ *schema.Builder) error { return nil }

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, r)
	assert.Equal(t, 0, r.Count())
	assert.Empty(t, r.GetAll())
}

func TestRegister_ValidName(t *testing.T) {
	r := NewRegistry()
	err := r.Register("20240215120000_create_users", &stubMigration{})
	require.NoError(t, err)
	assert.Equal(t, 1, r.Count())
}

func TestRegister_DuplicateName(t *testing.T) {
	r := NewRegistry()
	name := "20240215120000_create_users"
	require.NoError(t, r.Register(name, &stubMigration{}))

	err := r.Register(name, &stubMigration{})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrDuplicateMigration))
	assert.Contains(t, err.Error(), name)
}

func TestRegister_InvalidNames(t *testing.T) {
	tests := []struct {
		name string
		desc string
	}{
		{"", "empty string"},
		{"create_users", "no timestamp prefix"},
		{"2024021512_create_users", "short timestamp"},
		{"20240215120000", "no description"},
		{"20240215120000_", "empty description"},
		{"20240215120000_Create_users", "uppercase in description"},
		{"20240215120000_1create", "description starts with digit"},
		{"20240215120000_create-users", "hyphen in description"},
		{"20240215120000_create users", "space in description"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			r := NewRegistry()
			err := r.Register(tt.name, &stubMigration{})
			assert.Error(t, err)
			assert.True(t, errors.Is(err, ErrInvalidMigrationName), "expected ErrInvalidMigrationName for %q", tt.name)
		})
	}
}

func TestRegister_MaintainsSortedOrder(t *testing.T) {
	r := NewRegistry()
	// Register out of order.
	names := []string{
		"20240315000000_third",
		"20240115000000_first",
		"20240215000000_second",
	}
	for _, name := range names {
		require.NoError(t, r.Register(name, &stubMigration{}))
	}

	all := r.GetAll()
	require.Len(t, all, 3)
	assert.Equal(t, "20240115000000_first", all[0].Name)
	assert.Equal(t, "20240215000000_second", all[1].Name)
	assert.Equal(t, "20240315000000_third", all[2].Name)
}

func TestGet_Existing(t *testing.T) {
	r := NewRegistry()
	m := &stubMigration{}
	name := "20240215120000_create_users"
	require.NoError(t, r.Register(name, m))

	got, err := r.Get(name)
	require.NoError(t, err)
	assert.Same(t, m, got)
}

func TestGet_NotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.Get("20240215120000_nonexistent")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrMigrationNotFound))
}

func TestGetAll_ReturnsCopy(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register("20240115000000_first", &stubMigration{}))

	all1 := r.GetAll()
	all2 := r.GetAll()
	// Mutating the returned slice should not affect the registry.
	all1[0].Name = "mutated"
	assert.Equal(t, "20240115000000_first", all2[0].Name)
}

func TestCount(t *testing.T) {
	r := NewRegistry()
	assert.Equal(t, 0, r.Count())

	require.NoError(t, r.Register("20240115000000_a", &stubMigration{}))
	assert.Equal(t, 1, r.Count())

	require.NoError(t, r.Register("20240215000000_b", &stubMigration{}))
	assert.Equal(t, 2, r.Count())
}
