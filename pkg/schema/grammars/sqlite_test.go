package grammars

import (
	"errors"
	"strings"
	"testing"

	"github.com/andrianprasetya/go-migration/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSQLiteGrammar() *SQLiteGrammar {
	return NewSQLiteGrammar()
}

// --- Grammar interface compliance ---

func TestSQLiteGrammar_ImplementsGrammar(t *testing.T) {
	var _ schema.Grammar = (*SQLiteGrammar)(nil)
}

// --- CompileColumnType tests ---

func TestSQLite_CompileColumnType_String(t *testing.T) {
	g := newSQLiteGrammar()
	col := schema.ColumnDefinition{Name: "name", Type: schema.TypeString, Length: 100}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "TEXT", result)
}

func TestSQLite_CompileColumnType_Text(t *testing.T) {
	g := newSQLiteGrammar()
	col := schema.ColumnDefinition{Name: "body", Type: schema.TypeText}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "TEXT", result)
}

func TestSQLite_CompileColumnType_Integer(t *testing.T) {
	g := newSQLiteGrammar()
	col := schema.ColumnDefinition{Name: "age", Type: schema.TypeInteger}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "INTEGER", result)
}

func TestSQLite_CompileColumnType_IntegerAutoIncrementPK(t *testing.T) {
	g := newSQLiteGrammar()
	col := schema.ColumnDefinition{Name: "id", Type: schema.TypeInteger, IsAutoIncrement: true, IsPrimary: true}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "INTEGER PRIMARY KEY AUTOINCREMENT", result)
}

func TestSQLite_CompileColumnType_IntegerAutoIncrementNoPK(t *testing.T) {
	g := newSQLiteGrammar()
	// Auto-increment without primary key should just be INTEGER
	col := schema.ColumnDefinition{Name: "seq", Type: schema.TypeInteger, IsAutoIncrement: true}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "INTEGER", result)
}

func TestSQLite_CompileColumnType_BigInteger(t *testing.T) {
	g := newSQLiteGrammar()
	col := schema.ColumnDefinition{Name: "total", Type: schema.TypeBigInteger}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "INTEGER", result)
}

func TestSQLite_CompileColumnType_BigIntegerAutoIncrementPK(t *testing.T) {
	g := newSQLiteGrammar()
	col := schema.ColumnDefinition{Name: "id", Type: schema.TypeBigInteger, IsAutoIncrement: true, IsPrimary: true}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "INTEGER PRIMARY KEY AUTOINCREMENT", result)
}

func TestSQLite_CompileColumnType_Boolean(t *testing.T) {
	g := newSQLiteGrammar()
	col := schema.ColumnDefinition{Name: "active", Type: schema.TypeBoolean}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "INTEGER", result)
}

func TestSQLite_CompileColumnType_Timestamp(t *testing.T) {
	g := newSQLiteGrammar()
	col := schema.ColumnDefinition{Name: "created_at", Type: schema.TypeTimestamp}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "TEXT", result)
}

func TestSQLite_CompileColumnType_Date(t *testing.T) {
	g := newSQLiteGrammar()
	col := schema.ColumnDefinition{Name: "birth_date", Type: schema.TypeDate}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "TEXT", result)
}

func TestSQLite_CompileColumnType_Decimal(t *testing.T) {
	g := newSQLiteGrammar()
	col := schema.ColumnDefinition{Name: "price", Type: schema.TypeDecimal, Precision: 10, Scale: 2}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "REAL", result)
}

func TestSQLite_CompileColumnType_Float(t *testing.T) {
	g := newSQLiteGrammar()
	col := schema.ColumnDefinition{Name: "value", Type: schema.TypeFloat}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "REAL", result)
}

func TestSQLite_CompileColumnType_UUID(t *testing.T) {
	g := newSQLiteGrammar()
	col := schema.ColumnDefinition{Name: "external_id", Type: schema.TypeUUID}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "TEXT", result)
}

func TestSQLite_CompileColumnType_JSON(t *testing.T) {
	g := newSQLiteGrammar()
	col := schema.ColumnDefinition{Name: "metadata", Type: schema.TypeJSON}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "TEXT", result)
}

func TestSQLite_CompileColumnType_Binary(t *testing.T) {
	g := newSQLiteGrammar()
	col := schema.ColumnDefinition{Name: "data", Type: schema.TypeBinary}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "BLOB", result)
}

func TestSQLite_CompileColumnType_Unsupported(t *testing.T) {
	g := newSQLiteGrammar()
	col := schema.ColumnDefinition{Name: "bad", Type: schema.ColumnType(999)}
	_, err := g.CompileColumnType(col)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnsupportedType))
	assert.Contains(t, err.Error(), `"bad"`)
}

// --- CompileCreate tests ---

func TestSQLite_CompileCreate_SimpleTable(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("users")
	bp.ID()
	bp.String("name", 255)

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, `CREATE TABLE "users"`)
	assert.Contains(t, sql, `"id" INTEGER PRIMARY KEY AUTOINCREMENT`)
	assert.Contains(t, sql, `"name" TEXT NOT NULL`)
	// Auto-increment PK is inline, so no separate PRIMARY KEY clause
	assert.NotContains(t, sql, `PRIMARY KEY ("id")`)
}

func TestSQLite_CompileCreate_NoColumns(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("empty")

	_, err := g.CompileCreate(bp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no columns defined")
}

func TestSQLite_CompileCreate_AllColumnTypes(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("all_types")
	bp.String("col_string", 100)
	bp.Text("col_text")
	bp.Integer("col_int")
	bp.BigInteger("col_bigint")
	bp.Boolean("col_bool")
	bp.Timestamp("col_ts")
	bp.Date("col_date")
	bp.Decimal("col_dec", 8, 2)
	bp.Float("col_float")
	bp.UUID("col_uuid")
	bp.JSON("col_json")
	bp.Binary("col_bin")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	// All text-like types map to TEXT
	assert.Contains(t, sql, `"col_string" TEXT`)
	assert.Contains(t, sql, `"col_text" TEXT`)
	assert.Contains(t, sql, `"col_int" INTEGER`)
	assert.Contains(t, sql, `"col_bigint" INTEGER`)
	assert.Contains(t, sql, `"col_bool" INTEGER`)
	assert.Contains(t, sql, `"col_ts" TEXT`)
	assert.Contains(t, sql, `"col_date" TEXT`)
	assert.Contains(t, sql, `"col_dec" REAL`)
	assert.Contains(t, sql, `"col_float" REAL`)
	assert.Contains(t, sql, `"col_uuid" TEXT`)
	assert.Contains(t, sql, `"col_json" TEXT`)
	assert.Contains(t, sql, `"col_bin" BLOB`)
}

func TestSQLite_CompileCreate_NullableColumn(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("posts")
	bp.Text("body").Nullable()

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.NotContains(t, sql, "NOT NULL")
	assert.Contains(t, sql, `"body" TEXT`)
}

func TestSQLite_CompileCreate_DefaultValue(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("settings")
	bp.Boolean("active").Default(true)
	bp.String("status", 50).Default("pending")
	bp.Integer("count").Default(0)

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, "DEFAULT TRUE")
	assert.Contains(t, sql, "DEFAULT 'pending'")
	assert.Contains(t, sql, "DEFAULT 0")
}

func TestSQLite_CompileCreate_UniqueColumn(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("users")
	bp.String("email", 255).Unique()

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, "UNIQUE")
}

func TestSQLite_CompileCreate_WithIndexes(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("users")
	bp.String("name", 255)
	bp.String("email", 255)
	bp.Index("name")
	bp.UniqueIndex("email")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, `CONSTRAINT "uniq_users_email" UNIQUE ("email")`)
	assert.Contains(t, sql, `CREATE INDEX "idx_users_name" ON "users" ("name")`)
}

func TestSQLite_CompileCreate_WithForeignKey(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("posts")
	bp.BigInteger("user_id")
	bp.Foreign("user_id").References("id").On("users").OnDeleteAction("CASCADE").OnUpdateAction("SET NULL")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, `CONSTRAINT "fk_posts_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE SET NULL`)
}

func TestSQLite_CompileCreate_FullTable(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("users")
	bp.ID()
	bp.String("name", 255)
	bp.String("email", 255).Unique()
	bp.Boolean("active").Default(true)
	bp.Timestamps()
	bp.Index("name")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, `CREATE TABLE "users"`)
	assert.Contains(t, sql, `"id" INTEGER PRIMARY KEY AUTOINCREMENT`)
	assert.Contains(t, sql, `"name" TEXT NOT NULL`)
	assert.Contains(t, sql, `"email" TEXT NOT NULL UNIQUE`)
	assert.Contains(t, sql, `"active" INTEGER NOT NULL DEFAULT TRUE`)
	assert.Contains(t, sql, `"created_at" TEXT`)
	assert.Contains(t, sql, `"updated_at" TEXT`)
	assert.Contains(t, sql, `CREATE INDEX "idx_users_name" ON "users" ("name")`)
}

func TestSQLite_CompileCreate_CompositeIndex(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("users")
	bp.String("first_name", 100)
	bp.String("last_name", 100)
	bp.Index("first_name", "last_name")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Contains(t, sql, `CREATE INDEX "idx_users_first_name_last_name" ON "users" ("first_name", "last_name")`)
}

func TestSQLite_CompileCreate_Timestamps(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("posts")
	bp.String("title", 255)
	bp.Timestamps()

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, `"created_at" TEXT`)
	assert.Contains(t, sql, `"updated_at" TEXT`)
	// Timestamps should be nullable (no NOT NULL)
	parts := strings.Split(sql, ",")
	for _, part := range parts {
		if strings.Contains(part, "created_at") || strings.Contains(part, "updated_at") {
			assert.NotContains(t, part, "NOT NULL")
		}
	}
}

func TestSQLite_CompileCreate_NonPKAutoIncrement(t *testing.T) {
	// Auto-increment without primary key should just be INTEGER, no AUTOINCREMENT
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("counters")
	bp.Integer("seq").AutoIncrement()

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, `"seq" INTEGER NOT NULL`)
	assert.NotContains(t, sql, "AUTOINCREMENT")
}

func TestSQLite_CompileCreate_UnsignedIgnored(t *testing.T) {
	// SQLite ignores UNSIGNED — it should not appear in the output
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("items")
	bp.Integer("quantity").Unsigned()

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.NotContains(t, sql, "UNSIGNED")
	assert.Contains(t, sql, `"quantity" INTEGER NOT NULL`)
}

// --- CompileAlter tests ---

func TestSQLite_CompileAlter_AddColumn(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("users")
	bp.String("phone", 20).Nullable()

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `ALTER TABLE "users" ADD COLUMN "phone" TEXT`, stmts[0])
}

func TestSQLite_CompileAlter_DropColumn(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("users")
	bp.DropColumn("age")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `ALTER TABLE "users" DROP COLUMN "age"`, stmts[0])
}

func TestSQLite_CompileAlter_RenameColumn(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("users")
	bp.RenameColumn("name", "full_name")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `ALTER TABLE "users" RENAME COLUMN "name" TO "full_name"`, stmts[0])
}

func TestSQLite_CompileAlter_DropIndex(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("users")
	bp.DropIndex("idx_users_email")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `DROP INDEX "idx_users_email"`, stmts[0])
}

func TestSQLite_CompileAlter_AddIndex(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("users")
	bp.Index("name")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `CREATE INDEX "idx_users_name" ON "users" ("name")`, stmts[0])
}

func TestSQLite_CompileAlter_AddUniqueIndex(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("users")
	bp.UniqueIndex("email")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `CREATE UNIQUE INDEX "uniq_users_email" ON "users" ("email")`, stmts[0])
}

func TestSQLite_CompileAlter_AddForeignKey(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("posts")
	bp.Foreign("user_id").References("id").On("users").OnDeleteAction("CASCADE")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Contains(t, stmts[0], `ALTER TABLE "posts" ADD CONSTRAINT "fk_posts_user_id"`)
	assert.Contains(t, stmts[0], `FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE`)
}

func TestSQLite_CompileAlter_MultipleOperations(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("users")
	bp.String("phone", 20).Nullable()
	bp.DropColumn("age")
	bp.RenameColumn("name", "full_name")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 3)
	assert.Contains(t, stmts[0], "ADD COLUMN")
	assert.Contains(t, stmts[1], "DROP COLUMN")
	assert.Contains(t, stmts[2], "RENAME COLUMN")
}

// --- CompileDrop tests ---

func TestSQLite_CompileDrop(t *testing.T) {
	g := newSQLiteGrammar()
	assert.Equal(t, `DROP TABLE "users"`, g.CompileDrop("users"))
}

func TestSQLite_CompileDropIfExists(t *testing.T) {
	g := newSQLiteGrammar()
	assert.Equal(t, `DROP TABLE IF EXISTS "users"`, g.CompileDropIfExists("users"))
}

// --- CompileRename test ---

func TestSQLite_CompileRename(t *testing.T) {
	g := newSQLiteGrammar()
	assert.Equal(t, `ALTER TABLE "old_table" RENAME TO "new_table"`, g.CompileRename("old_table", "new_table"))
}

// --- CompileHasTable test ---

func TestSQLite_CompileHasTable(t *testing.T) {
	g := newSQLiteGrammar()
	sql := g.CompileHasTable("users")
	assert.Contains(t, sql, "sqlite_master")
	assert.Contains(t, sql, "type='table'")
	assert.Contains(t, sql, "name='users'")
}

// --- CompileHasColumn test ---

func TestSQLite_CompileHasColumn(t *testing.T) {
	g := newSQLiteGrammar()
	sql := g.CompileHasColumn("users", "email")
	assert.Contains(t, sql, "pragma_table_info('users')")
	assert.Contains(t, sql, "name='email'")
}

// --- CompileDropAllTables test ---

func TestSQLite_CompileDropAllTables(t *testing.T) {
	g := newSQLiteGrammar()
	sql := g.CompileDropAllTables()
	assert.Contains(t, sql, "sqlite_master")
	assert.Contains(t, sql, "type='table'")
	assert.Contains(t, sql, "NOT LIKE 'sqlite_%'")
}

// --- Default string escaping ---

func TestSQLite_CompileCreate_DefaultStringEscapesSingleQuotes(t *testing.T) {
	g := newSQLiteGrammar()
	bp := schema.NewBlueprint("test")
	bp.String("val", 100).Default("it's")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Contains(t, sql, "DEFAULT 'it''s'")
}
