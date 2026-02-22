package grammars

import (
	"errors"
	"strings"
	"testing"

	"github.com/andrianprasetya/go-migration/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newGrammar() *PostgresGrammar {
	return NewPostgresGrammar()
}

// --- CompileColumnType tests ---

func TestCompileColumnType_String(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "name", Type: schema.TypeString, Length: 100}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "VARCHAR(100)", result)
}

func TestCompileColumnType_StringDefaultLength(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "name", Type: schema.TypeString}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "VARCHAR(255)", result)
}

func TestCompileColumnType_Text(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "body", Type: schema.TypeText}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "TEXT", result)
}

func TestCompileColumnType_Integer(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "age", Type: schema.TypeInteger}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "INTEGER", result)
}

func TestCompileColumnType_IntegerAutoIncrement(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "id", Type: schema.TypeInteger, IsAutoIncrement: true}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "SERIAL", result)
}

func TestCompileColumnType_BigInteger(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "total", Type: schema.TypeBigInteger}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "BIGINT", result)
}

func TestCompileColumnType_BigIntegerAutoIncrement(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "id", Type: schema.TypeBigInteger, IsAutoIncrement: true}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "BIGSERIAL", result)
}

func TestCompileColumnType_Boolean(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "active", Type: schema.TypeBoolean}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "BOOLEAN", result)
}

func TestCompileColumnType_Timestamp(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "created_at", Type: schema.TypeTimestamp}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "TIMESTAMPTZ", result)
}

func TestCompileColumnType_Date(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "birth_date", Type: schema.TypeDate}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "DATE", result)
}

func TestCompileColumnType_Decimal(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "price", Type: schema.TypeDecimal, Precision: 10, Scale: 2}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "DECIMAL(10, 2)", result)
}

func TestCompileColumnType_DecimalDefaults(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "amount", Type: schema.TypeDecimal}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "DECIMAL(10, 0)", result)
}

func TestCompileColumnType_Float(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "value", Type: schema.TypeFloat}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "DOUBLE PRECISION", result)
}

func TestCompileColumnType_UUID(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "external_id", Type: schema.TypeUUID}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "UUID", result)
}

func TestCompileColumnType_JSON(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "metadata", Type: schema.TypeJSON}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "JSONB", result)
}

func TestCompileColumnType_Binary(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "data", Type: schema.TypeBinary}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "BYTEA", result)
}

func TestCompileColumnType_Unsupported(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "bad", Type: schema.ColumnType(999)}
	_, err := g.CompileColumnType(col)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnsupportedType))
	assert.Contains(t, err.Error(), `"bad"`)
}

// --- CompileCreate tests ---

func TestCompileCreate_SimpleTable(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("users")
	bp.ID()
	bp.String("name", 255)

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, `CREATE TABLE "users"`)
	assert.Contains(t, sql, `"id" BIGSERIAL`)
	assert.Contains(t, sql, `"name" VARCHAR(255) NOT NULL`)
	assert.Contains(t, sql, `PRIMARY KEY ("id")`)
}

func TestCompileCreate_NoColumns(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("empty")

	_, err := g.CompileCreate(bp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no columns defined")
}

func TestCompileCreate_AllColumnTypes(t *testing.T) {
	g := newGrammar()
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

	assert.Contains(t, sql, "VARCHAR(100)")
	assert.Contains(t, sql, "TEXT")
	assert.Contains(t, sql, "INTEGER")
	assert.Contains(t, sql, "BIGINT")
	assert.Contains(t, sql, "BOOLEAN")
	assert.Contains(t, sql, "TIMESTAMPTZ")
	assert.Contains(t, sql, "DATE")
	assert.Contains(t, sql, "DECIMAL(8, 2)")
	assert.Contains(t, sql, "DOUBLE PRECISION")
	assert.Contains(t, sql, "UUID")
	assert.Contains(t, sql, "JSONB")
	assert.Contains(t, sql, "BYTEA")
}

func TestCompileCreate_NullableColumn(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("posts")
	bp.Text("body").Nullable()

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	// Nullable columns should NOT have NOT NULL
	assert.NotContains(t, sql, "NOT NULL")
	assert.Contains(t, sql, `"body" TEXT`)
}

func TestCompileCreate_DefaultValue(t *testing.T) {
	g := newGrammar()
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

func TestCompileCreate_UniqueColumn(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("users")
	bp.String("email", 255).Unique()

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, "UNIQUE")
}

func TestCompileCreate_WithIndexes(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("users")
	bp.String("name", 255)
	bp.String("email", 255)
	bp.Index("name")
	bp.UniqueIndex("email")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	// Unique index should be inline as CONSTRAINT
	assert.Contains(t, sql, `CONSTRAINT "uniq_users_email" UNIQUE ("email")`)
	// Non-unique index should be a separate CREATE INDEX statement
	assert.Contains(t, sql, `CREATE INDEX "idx_users_name" ON "users" ("name")`)
}

func TestCompileCreate_WithForeignKey(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("posts")
	bp.BigInteger("user_id")
	bp.Foreign("user_id").References("id").On("users").OnDeleteAction("CASCADE").OnUpdateAction("SET NULL")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, `CONSTRAINT "fk_posts_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE SET NULL`)
}

func TestCompileCreate_FullTable(t *testing.T) {
	g := newGrammar()
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
	assert.Contains(t, sql, `"id" BIGSERIAL`)
	assert.Contains(t, sql, `"name" VARCHAR(255) NOT NULL`)
	assert.Contains(t, sql, `"email" VARCHAR(255) NOT NULL UNIQUE`)
	assert.Contains(t, sql, `"active" BOOLEAN NOT NULL DEFAULT TRUE`)
	assert.Contains(t, sql, `"created_at" TIMESTAMPTZ`)
	assert.Contains(t, sql, `"updated_at" TIMESTAMPTZ`)
	assert.Contains(t, sql, `PRIMARY KEY ("id")`)
	assert.Contains(t, sql, `CREATE INDEX "idx_users_name" ON "users" ("name")`)
}

// --- CompileAlter tests ---

func TestCompileAlter_AddColumn(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("users")
	bp.String("phone", 20).Nullable()

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `ALTER TABLE "users" ADD COLUMN "phone" VARCHAR(20)`, stmts[0])
}

func TestCompileAlter_DropColumn(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("users")
	bp.DropColumn("age")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `ALTER TABLE "users" DROP COLUMN "age"`, stmts[0])
}

func TestCompileAlter_RenameColumn(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("users")
	bp.RenameColumn("name", "full_name")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `ALTER TABLE "users" RENAME COLUMN "name" TO "full_name"`, stmts[0])
}

func TestCompileAlter_DropIndex(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("users")
	bp.DropIndex("idx_users_email")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `DROP INDEX "idx_users_email"`, stmts[0])
}

func TestCompileAlter_DropForeign(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("posts")
	bp.DropForeign("fk_posts_user_id")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `ALTER TABLE "posts" DROP CONSTRAINT "fk_posts_user_id"`, stmts[0])
}

func TestCompileAlter_AddIndex(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("users")
	bp.Index("name")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `CREATE INDEX "idx_users_name" ON "users" ("name")`, stmts[0])
}

func TestCompileAlter_AddUniqueIndex(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("users")
	bp.UniqueIndex("email")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `CREATE UNIQUE INDEX "uniq_users_email" ON "users" ("email")`, stmts[0])
}

func TestCompileAlter_AddForeignKey(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("posts")
	bp.Foreign("user_id").References("id").On("users").OnDeleteAction("CASCADE")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Contains(t, stmts[0], `ALTER TABLE "posts" ADD CONSTRAINT "fk_posts_user_id"`)
	assert.Contains(t, stmts[0], `FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE`)
}

func TestCompileAlter_MultipleOperations(t *testing.T) {
	g := newGrammar()
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

func TestCompileDrop(t *testing.T) {
	g := newGrammar()
	assert.Equal(t, `DROP TABLE "users"`, g.CompileDrop("users"))
}

func TestCompileDropIfExists(t *testing.T) {
	g := newGrammar()
	assert.Equal(t, `DROP TABLE IF EXISTS "users"`, g.CompileDropIfExists("users"))
}

// --- CompileRename test ---

func TestCompileRename(t *testing.T) {
	g := newGrammar()
	assert.Equal(t, `ALTER TABLE "old_table" RENAME TO "new_table"`, g.CompileRename("old_table", "new_table"))
}

// --- CompileHasTable test ---

func TestCompileHasTable(t *testing.T) {
	g := newGrammar()
	sql := g.CompileHasTable("users")
	assert.Contains(t, sql, "information_schema.tables")
	assert.Contains(t, sql, "table_name = 'users'")
	assert.Contains(t, sql, "table_schema = 'public'")
}

// --- CompileHasColumn test ---

func TestCompileHasColumn(t *testing.T) {
	g := newGrammar()
	sql := g.CompileHasColumn("users", "email")
	assert.Contains(t, sql, "information_schema.columns")
	assert.Contains(t, sql, "table_name = 'users'")
	assert.Contains(t, sql, "column_name = 'email'")
	assert.Contains(t, sql, "table_schema = 'public'")
}

// --- CompileDropAllTables test ---

func TestCompileDropAllTables(t *testing.T) {
	g := newGrammar()
	sql := g.CompileDropAllTables()
	assert.Contains(t, sql, "DROP SCHEMA public CASCADE")
	assert.Contains(t, sql, "CREATE SCHEMA public")
}

// --- Grammar interface compliance ---

func TestPostgresGrammar_ImplementsGrammar(t *testing.T) {
	var _ schema.Grammar = (*PostgresGrammar)(nil)
}

// --- Edge cases ---

func TestCompileCreate_CompositeIndex(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("users")
	bp.String("first_name", 100)
	bp.String("last_name", 100)
	bp.Index("first_name", "last_name")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Contains(t, sql, `CREATE INDEX "idx_users_first_name_last_name" ON "users" ("first_name", "last_name")`)
}

func TestCompileCreate_DefaultStringEscapesSingleQuotes(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("test")
	bp.String("val", 100).Default("it's")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Contains(t, sql, "DEFAULT 'it''s'")
}

func TestCompileCreate_Timestamps(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("posts")
	bp.String("title", 255)
	bp.Timestamps()

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	// Timestamps should be nullable (no NOT NULL)
	// Check that created_at and updated_at are present
	assert.Contains(t, sql, `"created_at" TIMESTAMPTZ`)
	assert.Contains(t, sql, `"updated_at" TIMESTAMPTZ`)
	// They should NOT have NOT NULL since they are nullable
	parts := strings.Split(sql, ",")
	for _, part := range parts {
		if strings.Contains(part, "created_at") || strings.Contains(part, "updated_at") {
			assert.NotContains(t, part, "NOT NULL")
		}
	}
}

// --- New column type compilation tests ---

func TestCompileColumnType_Enum(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "status", Type: schema.TypeEnum, AllowedValues: []string{"active", "inactive", "pending"}}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, `VARCHAR(255) CHECK ("status" IN ('active','inactive','pending'))`, result)
}

func TestCompileColumnType_EnumSingleValue(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "role", Type: schema.TypeEnum, AllowedValues: []string{"admin"}}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, `VARCHAR(255) CHECK ("role" IN ('admin'))`, result)
}

func TestCompileColumnType_EnumWithQuotes(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "label", Type: schema.TypeEnum, AllowedValues: []string{"it's", "they're"}}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, `VARCHAR(255) CHECK ("label" IN ('it''s','they''re'))`, result)
}

func TestCompileColumnType_EnumEmpty(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "status", Type: schema.TypeEnum, AllowedValues: []string{}}
	_, err := g.CompileColumnType(col)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "enum column requires at least one allowed value")
}

func TestCompileColumnType_EnumNilValues(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "status", Type: schema.TypeEnum}
	_, err := g.CompileColumnType(col)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "enum column requires at least one allowed value")
}

func TestCompileColumnType_Char(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "code", Type: schema.TypeChar, Length: 10}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "CHAR(10)", result)
}

func TestCompileColumnType_CharDefaultLength(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "code", Type: schema.TypeChar}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "CHAR(255)", result)
}

func TestCompileColumnType_CharZeroLength(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "code", Type: schema.TypeChar, Length: 0}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "CHAR(255)", result)
}

func TestCompileColumnType_LongText(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "content", Type: schema.TypeLongText}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "TEXT", result)
}

func TestCompileColumnType_MediumText(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "description", Type: schema.TypeMediumText}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "TEXT", result)
}

func TestCompileColumnType_TinyInt(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "priority", Type: schema.TypeTinyInt}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "SMALLINT", result)
}

func TestCompileColumnType_SmallInt(t *testing.T) {
	g := newGrammar()
	col := schema.ColumnDefinition{Name: "age", Type: schema.TypeSmallInt}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "SMALLINT", result)
}

// --- New column types in CompileCreate ---

func TestCompileCreate_WithEnumColumn(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("users")
	bp.String("name", 255)
	bp.Enum("status", []string{"active", "inactive"})

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Contains(t, sql, `VARCHAR(255) CHECK ("status" IN ('active','inactive'))`)
}

func TestCompileCreate_WithCharColumn(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("countries")
	bp.Char("code", 2)

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Contains(t, sql, `"code" CHAR(2)`)
}

func TestCompileCreate_WithNewTextTypes(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("articles")
	bp.LongText("body")
	bp.MediumText("summary")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Contains(t, sql, `"body" TEXT`)
	assert.Contains(t, sql, `"summary" TEXT`)
}

func TestCompileCreate_WithNewIntTypes(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("metrics")
	bp.TinyInt("priority")
	bp.SmallInt("score")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Contains(t, sql, `"priority" SMALLINT`)
	assert.Contains(t, sql, `"score" SMALLINT`)
}

// --- Fulltext and spatial index compilation tests ---

func TestCompileCreate_WithFulltextIndex(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("articles")
	bp.String("title", 255)
	bp.Text("body")
	bp.FulltextIndex("title")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Contains(t, sql, `CREATE INDEX "ft_articles_title" ON "articles" USING GIN (to_tsvector('english', "title"))`)
}

func TestCompileCreate_WithSpatialIndex(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("locations")
	bp.String("geom", 255)
	bp.SpatialIndex("geom")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Contains(t, sql, `CREATE INDEX "sp_locations_geom" ON "locations" USING GIST ("geom")`)
}

func TestCompileAlter_AddFulltextIndex(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("articles")
	bp.FulltextIndex("title")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `CREATE INDEX "ft_articles_title" ON "articles" USING GIN (to_tsvector('english', "title"))`, stmts[0])
}

func TestCompileAlter_AddSpatialIndex(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("locations")
	bp.SpatialIndex("geom")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `CREATE INDEX "sp_locations_geom" ON "locations" USING GIST ("geom")`, stmts[0])
}

func TestCompileCreate_WithFulltextAndRegularIndexes(t *testing.T) {
	g := newGrammar()
	bp := schema.NewBlueprint("posts")
	bp.String("title", 255)
	bp.String("slug", 255)
	bp.Index("slug")
	bp.FulltextIndex("title")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Contains(t, sql, `CREATE INDEX "idx_posts_slug" ON "posts" ("slug")`)
	assert.Contains(t, sql, `CREATE INDEX "ft_posts_title" ON "posts" USING GIN (to_tsvector('english', "title"))`)
}
