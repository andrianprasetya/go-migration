package generator

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// snakeCaseDescriptionGen generates valid snake_case migration/seeder descriptions.
func snakeCaseDescriptionGen() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		numParts := rapid.IntRange(1, 4).Draw(t, "numParts")
		parts := make([]string, numParts)
		for i := 0; i < numParts; i++ {
			parts[i] = rapid.StringMatching(`[a-z]{2,8}`).Draw(t, "part")
		}
		return strings.Join(parts, "_")
	})
}

// tableNameGen generates valid table names.
func tableNameGen() *rapid.Generator[string] {
	return rapid.StringMatching(`[a-z]{3,12}`)
}

// Feature: go-migration, Property 31: Generated migration files contain valid scaffolding
// **Validates: Requirements 16.1, 16.3**
//
// For any valid migration description, the generated file should contain a Go
// struct implementing the Migration interface with Up() and Down() stub methods,
// and the filename should match the pattern YYYYMMDDHHMMSS_description.go.
func TestProperty31_GeneratedMigrationFilesContainValidScaffolding(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		description := snakeCaseDescriptionGen().Draw(t, "description")

		dir, err := os.MkdirTemp("", "gen-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		fixedTime := time.Date(2024, 6, 15, 10, 30, 45, 0, time.UTC)
		gen := NewGenerator(dir)
		gen.nowFunc = func() time.Time { return fixedTime }

		outPath, err := gen.Migration(description, MigrationOptions{})
		require.NoError(t, err, "Migration generation should succeed")

		// Verify filename pattern: YYYYMMDDHHMMSS_description.go
		filename := filepath.Base(outPath)
		filenamePattern := regexp.MustCompile(`^\d{14}_` + regexp.QuoteMeta(description) + `\.go$`)
		assert.Regexp(t, filenamePattern, filename,
			"Filename should match YYYYMMDDHHMMSS_description.go pattern")

		// Verify timestamp prefix
		assert.True(t, strings.HasPrefix(filename, "20240615103045_"),
			"Filename should start with the correct timestamp")

		// Read generated content
		content, err := os.ReadFile(outPath)
		require.NoError(t, err)
		contentStr := string(content)

		// Verify struct name (PascalCase conversion of description)
		expectedStruct := toStructName(description)
		assert.Contains(t, contentStr, "type "+expectedStruct+" struct{}",
			"Should contain the struct definition")

		// Verify Up() method
		assert.Contains(t, contentStr, "func (m *"+expectedStruct+") Up(s *schema.Builder) error",
			"Should contain Up() method with correct receiver")

		// Verify Down() method
		assert.Contains(t, contentStr, "func (m *"+expectedStruct+") Down(s *schema.Builder) error",
			"Should contain Down() method with correct receiver")

		// Verify package declaration
		assert.Contains(t, contentStr, "package migrations",
			"Should declare package migrations")
	})
}

// Feature: go-migration, Property 32: Generated seeder files contain valid scaffolding
// **Validates: Requirements 16.2, 16.3**
//
// For any valid seeder description, the generated file should contain a Go
// struct implementing the Seeder interface with a Run() stub method, and the
// filename should match the pattern description_seeder.go.
func TestProperty32_GeneratedSeederFilesContainValidScaffolding(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		description := snakeCaseDescriptionGen().Draw(t, "description")

		dir, err := os.MkdirTemp("", "gen-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		gen := NewGenerator(dir)

		outPath, err := gen.Seeder(description)
		require.NoError(t, err, "Seeder generation should succeed")

		// Verify filename pattern: description_seeder.go
		filename := filepath.Base(outPath)
		expectedFilename := description + "_seeder.go"
		assert.Equal(t, expectedFilename, filename,
			"Filename should match description_seeder.go pattern")

		// Read generated content
		content, err := os.ReadFile(outPath)
		require.NoError(t, err)
		contentStr := string(content)

		// Verify struct name (PascalCase conversion of description)
		expectedStruct := toStructName(description)
		assert.Contains(t, contentStr, "type "+expectedStruct+" struct{}",
			"Should contain the struct definition")

		// Verify Run() method
		assert.Contains(t, contentStr, "func (s *"+expectedStruct+") Run(db *sql.DB) error",
			"Should contain Run() method with correct receiver")

		// Verify package declaration
		assert.Contains(t, contentStr, "package seeders",
			"Should declare package seeders")
	})
}

// Feature: go-migration, Property 33: Create flag pre-populates schema create call
// **Validates: Requirements 16.4**
//
// For any table name provided via --create flag, the generated migration Up()
// method should contain a Schema_Builder Create() call with that table name.
func TestProperty33_CreateFlagPrePopulatesSchemaCreateCall(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		description := snakeCaseDescriptionGen().Draw(t, "description")
		tableName := tableNameGen().Draw(t, "tableName")

		dir, err := os.MkdirTemp("", "gen-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		gen := NewGenerator(dir)
		gen.nowFunc = func() time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) }

		outPath, err := gen.Migration(description, MigrationOptions{
			CreateTable: tableName,
		})
		require.NoError(t, err, "Migration generation with --create should succeed")

		content, err := os.ReadFile(outPath)
		require.NoError(t, err)
		contentStr := string(content)

		// Up() should contain s.Create("tableName", ...)
		assert.Contains(t, contentStr, `s.Create("`+tableName+`"`,
			"Up() should contain Schema_Builder Create() call with the table name")

		// Down() should contain s.Drop("tableName")
		assert.Contains(t, contentStr, `s.Drop("`+tableName+`"`,
			"Down() should contain Schema_Builder Drop() call with the table name")

		// Verify struct is still correct
		expectedStruct := toStructName(description)
		assert.Contains(t, contentStr, "type "+expectedStruct+" struct{}",
			"Should contain the struct definition")
	})
}

// Feature: go-migration, Property 34: Table flag pre-populates schema alter call
// **Validates: Requirements 16.5**
//
// For any table name provided via --table flag, the generated migration Up()
// method should contain a Schema_Builder Alter() call with that table name.
func TestProperty34_TableFlagPrePopulatesSchemaAlterCall(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		description := snakeCaseDescriptionGen().Draw(t, "description")
		tableName := tableNameGen().Draw(t, "tableName")

		dir, err := os.MkdirTemp("", "gen-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		gen := NewGenerator(dir)
		gen.nowFunc = func() time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) }

		outPath, err := gen.Migration(description, MigrationOptions{
			AlterTable: tableName,
		})
		require.NoError(t, err, "Migration generation with --table should succeed")

		content, err := os.ReadFile(outPath)
		require.NoError(t, err)
		contentStr := string(content)

		// Up() should contain s.Alter("tableName", ...)
		assert.Contains(t, contentStr, `s.Alter("`+tableName+`"`,
			"Up() should contain Schema_Builder Alter() call with the table name")

		// Down() should also contain s.Alter("tableName", ...) for reversal
		// Count occurrences - both Up and Down should reference the table
		alterCount := strings.Count(contentStr, `s.Alter("`+tableName+`"`)
		assert.Equal(t, 2, alterCount,
			"Both Up() and Down() should contain Alter() calls with the table name")

		// Verify struct is still correct
		expectedStruct := toStructName(description)
		assert.Contains(t, contentStr, "type "+expectedStruct+" struct{}",
			"Should contain the struct definition")
	})
}
