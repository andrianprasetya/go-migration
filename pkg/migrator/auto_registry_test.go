package migrator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutoRegister_ValidName(t *testing.T) {
	defer ResetAutoRegistry()

	AutoRegister("20240101000001_create_users", &stubMigration{})

	result := GetAutoRegistered()
	assert.Len(t, result, 1)
	assert.Equal(t, "20240101000001_create_users", result[0].Name)
}

func TestAutoRegister_SortedOrder(t *testing.T) {
	defer ResetAutoRegistry()

	// Register out of order
	AutoRegister("20240101000003_create_posts", &stubMigration{})
	AutoRegister("20240101000001_create_users", &stubMigration{})
	AutoRegister("20240101000002_create_comments", &stubMigration{})

	result := GetAutoRegistered()
	assert.Len(t, result, 3)
	assert.Equal(t, "20240101000001_create_users", result[0].Name)
	assert.Equal(t, "20240101000002_create_comments", result[1].Name)
	assert.Equal(t, "20240101000003_create_posts", result[2].Name)
}

func TestAutoRegister_InvalidNamePanics(t *testing.T) {
	defer ResetAutoRegistry()

	assert.PanicsWithValue(t,
		`AutoRegister: migration name "bad-name" is invalid (expected YYYYMMDDHHMMSS_description or YYYY_MM_DD_HHMMSS_RRRR_description)`,
		func() { AutoRegister("bad-name", &stubMigration{}) },
	)
}

func TestAutoRegister_DuplicateNamePanics(t *testing.T) {
	defer ResetAutoRegistry()

	AutoRegister("20240101000001_create_users", &stubMigration{})

	assert.PanicsWithValue(t,
		`AutoRegister: duplicate migration name "20240101000001_create_users"`,
		func() { AutoRegister("20240101000001_create_users", &stubMigration{}) },
	)
}

func TestGetAutoRegistered_ReturnsCopy(t *testing.T) {
	defer ResetAutoRegistry()

	AutoRegister("20240101000001_create_users", &stubMigration{})

	result1 := GetAutoRegistered()
	result2 := GetAutoRegistered()

	// Modifying the returned slice should not affect the registry
	result1[0].Name = "modified"
	assert.Equal(t, "20240101000001_create_users", result2[0].Name)
}

func TestGetAutoRegistered_EmptyRegistry(t *testing.T) {
	defer ResetAutoRegistry()

	result := GetAutoRegistered()
	assert.Empty(t, result)
	assert.NotNil(t, result)
}

func TestResetAutoRegistry(t *testing.T) {
	AutoRegister("20240101000001_create_users", &stubMigration{})
	assert.Len(t, GetAutoRegistered(), 1)

	ResetAutoRegistry()
	assert.Empty(t, GetAutoRegistered())
}
