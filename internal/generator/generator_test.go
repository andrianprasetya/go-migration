package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixedTime returns a generator with a fixed timestamp for deterministic filenames.
func fixedTimeGenerator(t *testing.T) (*Generator, string) {
	t.Helper()
	dir := t.TempDir()
	g := NewGenerator(dir)
	g.nowFunc = func() time.Time {
		return time.Date(2024, 7, 15, 10, 30, 45, 0, time.UTC)
	}
	return g, dir
}

func TestMigration_BasicFile(t *testing.T) {
	g, dir := fixedTimeGenerator(t)

	path, err := g.Migration("create_users", MigrationOptions{})
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(dir, "20240715103045_create_users.go"), path)

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	src := string(content)
	assert.Contains(t, src, "type CreateUsers struct{}")
	assert.Contains(t, src, "func (m *CreateUsers) Up(s *schema.Builder) error")
	assert.Contains(t, src, "func (m *CreateUsers) Down(s *schema.Builder) error")
	assert.Contains(t, src, "package migrations")
}

func TestMigration_CreateFlag(t *testing.T) {
	g, _ := fixedTimeGenerator(t)

	path, err := g.Migration("create_orders", MigrationOptions{CreateTable: "orders"})
	require.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	src := string(content)
	assert.Contains(t, src, "type CreateOrders struct{}")
	assert.Contains(t, src, `s.Create("orders"`)
	assert.Contains(t, src, `s.Drop("orders"`)
	assert.Contains(t, src, "bp.ID()")
	assert.Contains(t, src, "bp.Timestamps()")
}

func TestMigration_TableFlag(t *testing.T) {
	g, _ := fixedTimeGenerator(t)

	path, err := g.Migration("add_email_to_users", MigrationOptions{AlterTable: "users"})
	require.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	src := string(content)
	assert.Contains(t, src, "type AddEmailToUsers struct{}")
	assert.Contains(t, src, `s.Alter("users"`)
}

func TestSeeder_BasicFile(t *testing.T) {
	g, dir := fixedTimeGenerator(t)

	path, err := g.Seeder("user")
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(dir, "user_seeder.go"), path)

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	src := string(content)
	assert.Contains(t, src, "type User struct{}")
	assert.Contains(t, src, "func (s *User) Run(db *sql.DB) error")
	assert.Contains(t, src, "package seeders")
}

func TestMigration_FilenameFormat(t *testing.T) {
	g, _ := fixedTimeGenerator(t)

	path, err := g.Migration("create_products", MigrationOptions{})
	require.NoError(t, err)

	filename := filepath.Base(path)
	// Should match YYYYMMDDHHMMSS_description.go
	assert.True(t, strings.HasSuffix(filename, ".go"))
	assert.Equal(t, "20240715103045_create_products.go", filename)
}

func TestSeeder_FilenameFormat(t *testing.T) {
	g, _ := fixedTimeGenerator(t)

	path, err := g.Seeder("product")
	require.NoError(t, err)

	filename := filepath.Base(path)
	assert.Equal(t, "product_seeder.go", filename)
}

func TestMigration_StructAndMethods(t *testing.T) {
	g, _ := fixedTimeGenerator(t)

	path, err := g.Migration("add_indexes", MigrationOptions{})
	require.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	src := string(content)
	// Struct implements Migration interface
	assert.Contains(t, src, "type AddIndexes struct{}")
	assert.Contains(t, src, "func (m *AddIndexes) Up(s *schema.Builder) error")
	assert.Contains(t, src, "func (m *AddIndexes) Down(s *schema.Builder) error")
	// Imports schema package
	assert.Contains(t, src, `"github.com/andrianprasetya/go-migration/pkg/schema"`)
}

func TestToStructName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"create_users", "CreateUsers"},
		{"add_email_to_users", "AddEmailToUsers"},
		{"user", "User"},
		{"create_order_items", "CreateOrderItems"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, toStructName(tt.input))
		})
	}
}
