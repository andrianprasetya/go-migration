package seeder

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSeeder is a simple Seeder implementation for testing.
type mockSeeder struct {
	name string
}

func (m *mockSeeder) Run(db *sql.DB) error {
	return nil
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, r)
	assert.Empty(t, r.GetAll())
}

func TestRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	s := &mockSeeder{name: "users"}

	err := r.Register("users", s)
	require.NoError(t, err)

	got, err := r.Get("users")
	require.NoError(t, err)
	assert.Equal(t, s, got)
}

func TestRegisterDuplicateNameReturnsError(t *testing.T) {
	r := NewRegistry()
	s1 := &mockSeeder{name: "users"}
	s2 := &mockSeeder{name: "users2"}

	err := r.Register("users", s1)
	require.NoError(t, err)

	err = r.Register("users", s2)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrDuplicateSeeder))
	assert.Contains(t, err.Error(), "users")
}

func TestRegisterEmptyNameReturnsError(t *testing.T) {
	r := NewRegistry()
	s := &mockSeeder{name: "test"}

	tests := []struct {
		name string
	}{
		{""},
		{" "},
		{"  "},
		{"\t"},
	}

	for _, tc := range tests {
		err := r.Register(tc.name, s)
		require.Error(t, err, "expected error for name %q", tc.name)
		assert.True(t, errors.Is(err, ErrInvalidSeederName))
	}
}

func TestGetNonExistentReturnsError(t *testing.T) {
	r := NewRegistry()

	_, err := r.Get("nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSeederNotFound))
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestGetAllReturnsAllRegistered(t *testing.T) {
	r := NewRegistry()
	s1 := &mockSeeder{name: "users"}
	s2 := &mockSeeder{name: "posts"}
	s3 := &mockSeeder{name: "comments"}

	require.NoError(t, r.Register("users", s1))
	require.NoError(t, r.Register("posts", s2))
	require.NoError(t, r.Register("comments", s3))

	all := r.GetAll()
	assert.Len(t, all, 3)
	assert.Equal(t, s1, all["users"])
	assert.Equal(t, s2, all["posts"])
	assert.Equal(t, s3, all["comments"])
}

func TestGetAllReturnsCopy(t *testing.T) {
	r := NewRegistry()
	s := &mockSeeder{name: "users"}
	require.NoError(t, r.Register("users", s))

	all := r.GetAll()
	// Mutating the returned map should not affect the registry.
	delete(all, "users")

	got, err := r.Get("users")
	require.NoError(t, err)
	assert.Equal(t, s, got)
}
