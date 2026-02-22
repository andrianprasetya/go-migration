package schema_test

import (
	"testing"

	"github.com/andrianprasetya/go-migration/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// Feature: library-improvements, Property 1: New column type blueprint methods preserve type and attributes
// **Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5, 1.6**
//
// For any new column type (Enum, Char, LongText, MediumText, TinyInt, SmallInt) and any valid
// arguments (name string, length int, allowed values), calling the corresponding Blueprint method
// SHALL add a ColumnDefinition with the correct ColumnType, and for Enum the AllowedValues SHALL
// match, and for Char the Length SHALL match.
func TestProperty1_NewColumnTypeBlueprintMethodsPreserveTypeAndAttributes(t *testing.T) {
	// Generator for valid column names (non-empty, alphanumeric with underscores)
	colNameGen := rapid.StringMatching(`[a-z][a-z0-9_]{0,19}`)

	t.Run("Enum preserves type and allowed values", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			bp := schema.NewBlueprint("test_table")
			name := colNameGen.Draw(t, "name")
			numValues := rapid.IntRange(1, 10).Draw(t, "numValues")
			allowed := make([]string, numValues)
			for i := range allowed {
				allowed[i] = rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9_]{0,9}`).Draw(t, "allowedValue")
			}

			col := bp.Enum(name, allowed)

			require.NotNil(t, col)
			assert.Equal(t, name, col.Name)
			assert.Equal(t, schema.TypeEnum, col.Type)
			assert.Equal(t, allowed, col.AllowedValues)

			// Verify it was added to the blueprint
			cols := bp.Columns()
			require.Len(t, cols, 1)
			assert.Equal(t, schema.TypeEnum, cols[0].Type)
			assert.Equal(t, allowed, cols[0].AllowedValues)
		})
	})

	t.Run("Char preserves type and length", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			bp := schema.NewBlueprint("test_table")
			name := colNameGen.Draw(t, "name")
			length := rapid.IntRange(1, 255).Draw(t, "length")

			col := bp.Char(name, length)

			require.NotNil(t, col)
			assert.Equal(t, name, col.Name)
			assert.Equal(t, schema.TypeChar, col.Type)
			assert.Equal(t, length, col.Length)

			cols := bp.Columns()
			require.Len(t, cols, 1)
			assert.Equal(t, schema.TypeChar, cols[0].Type)
			assert.Equal(t, length, cols[0].Length)
		})
	})

	t.Run("LongText preserves type", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			bp := schema.NewBlueprint("test_table")
			name := colNameGen.Draw(t, "name")

			col := bp.LongText(name)

			require.NotNil(t, col)
			assert.Equal(t, name, col.Name)
			assert.Equal(t, schema.TypeLongText, col.Type)

			cols := bp.Columns()
			require.Len(t, cols, 1)
			assert.Equal(t, schema.TypeLongText, cols[0].Type)
		})
	})

	t.Run("MediumText preserves type", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			bp := schema.NewBlueprint("test_table")
			name := colNameGen.Draw(t, "name")

			col := bp.MediumText(name)

			require.NotNil(t, col)
			assert.Equal(t, name, col.Name)
			assert.Equal(t, schema.TypeMediumText, col.Type)

			cols := bp.Columns()
			require.Len(t, cols, 1)
			assert.Equal(t, schema.TypeMediumText, cols[0].Type)
		})
	})

	t.Run("TinyInt preserves type", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			bp := schema.NewBlueprint("test_table")
			name := colNameGen.Draw(t, "name")

			col := bp.TinyInt(name)

			require.NotNil(t, col)
			assert.Equal(t, name, col.Name)
			assert.Equal(t, schema.TypeTinyInt, col.Type)

			cols := bp.Columns()
			require.Len(t, cols, 1)
			assert.Equal(t, schema.TypeTinyInt, cols[0].Type)
		})
	})

	t.Run("SmallInt preserves type", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			bp := schema.NewBlueprint("test_table")
			name := colNameGen.Draw(t, "name")

			col := bp.SmallInt(name)

			require.NotNil(t, col)
			assert.Equal(t, name, col.Name)
			assert.Equal(t, schema.TypeSmallInt, col.Type)

			cols := bp.Columns()
			require.Len(t, cols, 1)
			assert.Equal(t, schema.TypeSmallInt, cols[0].Type)
		})
	})
}

// Feature: library-improvements, Property 3: New index type blueprint methods set correct IndexType
// **Validates: Requirements 2.1, 2.2**
//
// For any set of column names, calling FulltextIndex(columns...) SHALL produce an IndexDefinition
// with Type == IndexFulltext, and calling SpatialIndex(columns...) SHALL produce an IndexDefinition
// with Type == IndexSpatial. The columns in the definition SHALL match the input columns.
func TestProperty3_NewIndexTypeBlueprintMethodsSetCorrectIndexType(t *testing.T) {
	// Generator for valid column names
	colNameGen := rapid.StringMatching(`[a-z][a-z0-9_]{0,19}`)

	// Generator for a non-empty slice of unique column names
	colsGen := rapid.Custom(func(t *rapid.T) []string {
		n := rapid.IntRange(1, 5).Draw(t, "numCols")
		seen := make(map[string]bool)
		cols := make([]string, 0, n)
		for len(cols) < n {
			c := colNameGen.Draw(t, "col")
			if !seen[c] {
				seen[c] = true
				cols = append(cols, c)
			}
		}
		return cols
	})

	t.Run("FulltextIndex sets IndexFulltext and preserves columns", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			tableName := colNameGen.Draw(t, "table")
			columns := colsGen.Draw(t, "columns")

			bp := schema.NewBlueprint(tableName)
			idx := bp.FulltextIndex(columns...)

			require.NotNil(t, idx)
			assert.Equal(t, schema.IndexFulltext, idx.Type)
			assert.Equal(t, columns, idx.Columns)

			// Verify it was added to the blueprint
			indexes := bp.Indexes()
			require.Len(t, indexes, 1)
			assert.Equal(t, schema.IndexFulltext, indexes[0].Type)
			assert.Equal(t, columns, indexes[0].Columns)
		})
	})

	t.Run("SpatialIndex sets IndexSpatial and preserves columns", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			tableName := colNameGen.Draw(t, "table")
			columns := colsGen.Draw(t, "columns")

			bp := schema.NewBlueprint(tableName)
			idx := bp.SpatialIndex(columns...)

			require.NotNil(t, idx)
			assert.Equal(t, schema.IndexSpatial, idx.Type)
			assert.Equal(t, columns, idx.Columns)

			// Verify it was added to the blueprint
			indexes := bp.Indexes()
			require.Len(t, indexes, 1)
			assert.Equal(t, schema.IndexSpatial, indexes[0].Type)
			assert.Equal(t, columns, indexes[0].Columns)
		})
	})
}
