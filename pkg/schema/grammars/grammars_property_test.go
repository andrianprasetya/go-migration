package grammars

import (
	"strings"
	"testing"

	"github.com/andrianprasetya/go-migration/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// allowedValueGen generates a non-empty allowed value string suitable for enum values.
// Avoids single quotes to keep SQL quoting straightforward.
func allowedValueGen() *rapid.Generator[string] {
	return rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9_]{0,19}`)
}

// allowedValuesGen generates a non-empty slice of unique allowed values for enum columns.
func allowedValuesGen() *rapid.Generator[[]string] {
	return rapid.Custom(func(t *rapid.T) []string {
		count := rapid.IntRange(1, 10).Draw(t, "numValues")
		seen := make(map[string]bool)
		values := make([]string, 0, count)
		for len(values) < count {
			v := allowedValueGen().Filter(func(s string) bool {
				return !seen[s]
			}).Draw(t, "allowedValue")
			seen[v] = true
			values = append(values, v)
		}
		return values
	})
}

// Feature: library-improvements, Property 2: Enum compilation produces grammar-appropriate SQL with all allowed values
// **Validates: Requirements 1.7, 1.8, 1.9**
func TestProperty2_EnumCompilationProducesGrammarAppropriateSQL(t *testing.T) {
	grammars := allGrammars()

	for _, g := range grammars {
		t.Run(g.name, func(t *testing.T) {
			rapid.Check(t, func(t *rapid.T) {
				colName := columnNameGen().Draw(t, "colName")
				allowedValues := allowedValuesGen().Draw(t, "allowedValues")

				col := schema.ColumnDefinition{
					Name:          colName,
					Type:          schema.TypeEnum,
					AllowedValues: allowedValues,
				}

				sql, err := g.grammar.CompileColumnType(col)
				require.NoError(t, err, "CompileColumnType should not error for valid enum with grammar %s", g.name)

				// Every allowed value must appear in the output SQL
				for _, v := range allowedValues {
					assert.Contains(t, sql, v,
						"SQL output for grammar %s should contain allowed value %q, got: %s", g.name, v, sql)
				}

				// Grammar-specific structural checks
				switch g.name {
				case "postgres":
					assert.Contains(t, sql, "VARCHAR(255)",
						"PostgreSQL enum should contain VARCHAR(255), got: %s", sql)
					assert.Contains(t, sql, "CHECK",
						"PostgreSQL enum should contain CHECK constraint, got: %s", sql)
					assert.Contains(t, sql, colName,
						"PostgreSQL enum CHECK should reference column name %q, got: %s", colName, sql)

				case "mysql":
					assert.True(t, strings.HasPrefix(sql, "ENUM("),
						"MySQL enum should start with ENUM(, got: %s", sql)

				case "sqlite":
					assert.True(t, strings.HasPrefix(sql, "TEXT CHECK"),
						"SQLite enum should start with TEXT CHECK, got: %s", sql)
					assert.Contains(t, sql, colName,
						"SQLite enum CHECK should reference column name %q, got: %s", colName, sql)
				}
			})
		})
	}
}

// columnsGen generates a non-empty slice of unique column names for index testing.
func columnsGen() *rapid.Generator[[]string] {
	return rapid.Custom(func(t *rapid.T) []string {
		count := rapid.IntRange(1, 5).Draw(t, "numColumns")
		seen := make(map[string]bool)
		cols := make([]string, 0, count)
		for len(cols) < count {
			c := columnNameGen().Filter(func(s string) bool {
				return !seen[s]
			}).Draw(t, "col")
			seen[c] = true
			cols = append(cols, c)
		}
		return cols
	})
}

// Feature: library-improvements, Property 4: Fulltext and spatial index compilation produces grammar-appropriate SQL
// **Validates: Requirements 2.3, 2.4, 2.5, 2.7, 2.8**
func TestProperty4_FulltextSpatialIndexCompilationProducesGrammarAppropriateSQL(t *testing.T) {
	pgGrammar := NewPostgresGrammar()
	myGrammar := NewMySQLGrammar()

	t.Run("Fulltext_Postgres_CompileCreate", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			tableName := tableNameGen().Draw(t, "tableName")
			cols := columnsGen().Draw(t, "columns")

			bp := schema.NewBlueprint(tableName)
			bp.ID()
			for _, c := range cols {
				bp.Text(c)
			}
			bp.FulltextIndex(cols...)

			sql, err := pgGrammar.CompileCreate(bp)
			require.NoError(t, err)
			assert.Contains(t, sql, "USING GIN")
			assert.Contains(t, sql, "to_tsvector")
		})
	})

	t.Run("Fulltext_Postgres_CompileAlter", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			tableName := tableNameGen().Draw(t, "tableName")
			cols := columnsGen().Draw(t, "columns")

			bp := schema.NewBlueprint(tableName)
			bp.FulltextIndex(cols...)

			stmts, err := pgGrammar.CompileAlter(bp)
			require.NoError(t, err)

			joined := strings.Join(stmts, "; ")
			assert.Contains(t, joined, "USING GIN")
			assert.Contains(t, joined, "to_tsvector")
		})
	})

	t.Run("Fulltext_MySQL_CompileCreate", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			tableName := tableNameGen().Draw(t, "tableName")
			cols := columnsGen().Draw(t, "columns")

			bp := schema.NewBlueprint(tableName)
			bp.ID()
			for _, c := range cols {
				bp.Text(c)
			}
			bp.FulltextIndex(cols...)

			sql, err := myGrammar.CompileCreate(bp)
			require.NoError(t, err)
			assert.Contains(t, sql, "FULLTEXT")
		})
	})

	t.Run("Fulltext_MySQL_CompileAlter", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			tableName := tableNameGen().Draw(t, "tableName")
			cols := columnsGen().Draw(t, "columns")

			bp := schema.NewBlueprint(tableName)
			bp.FulltextIndex(cols...)

			stmts, err := myGrammar.CompileAlter(bp)
			require.NoError(t, err)

			joined := strings.Join(stmts, "; ")
			assert.Contains(t, joined, "FULLTEXT")
		})
	})

	t.Run("Spatial_Postgres_CompileCreate", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			tableName := tableNameGen().Draw(t, "tableName")
			cols := columnsGen().Draw(t, "columns")

			bp := schema.NewBlueprint(tableName)
			bp.ID()
			for _, c := range cols {
				bp.Text(c)
			}
			bp.SpatialIndex(cols...)

			sql, err := pgGrammar.CompileCreate(bp)
			require.NoError(t, err)
			assert.Contains(t, sql, "USING GIST")
		})
	})

	t.Run("Spatial_Postgres_CompileAlter", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			tableName := tableNameGen().Draw(t, "tableName")
			cols := columnsGen().Draw(t, "columns")

			bp := schema.NewBlueprint(tableName)
			bp.SpatialIndex(cols...)

			stmts, err := pgGrammar.CompileAlter(bp)
			require.NoError(t, err)

			joined := strings.Join(stmts, "; ")
			assert.Contains(t, joined, "USING GIST")
		})
	})

	t.Run("Spatial_MySQL_CompileCreate", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			tableName := tableNameGen().Draw(t, "tableName")
			cols := columnsGen().Draw(t, "columns")

			bp := schema.NewBlueprint(tableName)
			bp.ID()
			for _, c := range cols {
				bp.Text(c)
			}
			bp.SpatialIndex(cols...)

			sql, err := myGrammar.CompileCreate(bp)
			require.NoError(t, err)
			assert.Contains(t, sql, "SPATIAL")
		})
	})

	t.Run("Spatial_MySQL_CompileAlter", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			tableName := tableNameGen().Draw(t, "tableName")
			cols := columnsGen().Draw(t, "columns")

			bp := schema.NewBlueprint(tableName)
			bp.SpatialIndex(cols...)

			stmts, err := myGrammar.CompileAlter(bp)
			require.NoError(t, err)

			joined := strings.Join(stmts, "; ")
			assert.Contains(t, joined, "SPATIAL")
		})
	})
}

// Feature: library-improvements, Property 5: SQLite grammar rejects fulltext and spatial indexes
// **Validates: Requirements 2.6, 2.9**
func TestProperty5_SQLiteGrammarRejectsFulltextAndSpatialIndexes(t *testing.T) {
	sqliteGrammar := NewSQLiteGrammar()

	t.Run("Fulltext_CompileCreate", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			tableName := tableNameGen().Draw(t, "tableName")
			cols := columnsGen().Draw(t, "columns")

			bp := schema.NewBlueprint(tableName)
			bp.ID()
			for _, c := range cols {
				bp.Text(c)
			}
			bp.FulltextIndex(cols...)

			_, err := sqliteGrammar.CompileCreate(bp)
			assert.Error(t, err, "SQLite CompileCreate should reject fulltext indexes")
		})
	})

	t.Run("Fulltext_CompileAlter", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			tableName := tableNameGen().Draw(t, "tableName")
			cols := columnsGen().Draw(t, "columns")

			bp := schema.NewBlueprint(tableName)
			bp.FulltextIndex(cols...)

			_, err := sqliteGrammar.CompileAlter(bp)
			assert.Error(t, err, "SQLite CompileAlter should reject fulltext indexes")
		})
	})

	t.Run("Spatial_CompileCreate", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			tableName := tableNameGen().Draw(t, "tableName")
			cols := columnsGen().Draw(t, "columns")

			bp := schema.NewBlueprint(tableName)
			bp.ID()
			for _, c := range cols {
				bp.Text(c)
			}
			bp.SpatialIndex(cols...)

			_, err := sqliteGrammar.CompileCreate(bp)
			assert.Error(t, err, "SQLite CompileCreate should reject spatial indexes")
		})
	})

	t.Run("Spatial_CompileAlter", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			tableName := tableNameGen().Draw(t, "tableName")
			cols := columnsGen().Draw(t, "columns")

			bp := schema.NewBlueprint(tableName)
			bp.SpatialIndex(cols...)

			_, err := sqliteGrammar.CompileAlter(bp)
			assert.Error(t, err, "SQLite CompileAlter should reject spatial indexes")
		})
	})
}
