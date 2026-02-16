package grammars

import (
	"errors"
	"strings"
	"testing"

	"github.com/andrianprasetya/go-migration/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMySQLGrammar() *MySQLGrammar {
	return NewMySQLGrammar()
}

// --- CompileColumnType tests ---

func TestMySQL_CompileColumnType_String(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "name", Type: schema.TypeString, Length: 100}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "VARCHAR(100)", result)
}

func TestMySQL_CompileColumnType_StringDefaultLength(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "name", Type: schema.TypeString}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "VARCHAR(255)", result)
}

func TestMySQL_CompileColumnType_Text(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "body", Type: schema.TypeText}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "TEXT", result)
}

func TestMySQL_CompileColumnType_Integer(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "age", Type: schema.TypeInteger}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "INT", result)
}

func TestMySQL_CompileColumnType_IntegerUnsigned(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "count", Type: schema.TypeInteger, IsUnsigned: true}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "INT UNSIGNED", result)
}

func TestMySQL_CompileColumnType_IntegerAutoIncrement(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "id", Type: schema.TypeInteger, IsAutoIncrement: true}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "INT AUTO_INCREMENT", result)
}

func TestMySQL_CompileColumnType_IntegerUnsignedAutoIncrement(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "id", Type: schema.TypeInteger, IsUnsigned: true, IsAutoIncrement: true}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "INT UNSIGNED AUTO_INCREMENT", result)
}

func TestMySQL_CompileColumnType_BigInteger(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "total", Type: schema.TypeBigInteger}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "BIGINT", result)
}

func TestMySQL_CompileColumnType_BigIntegerUnsigned(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "total", Type: schema.TypeBigInteger, IsUnsigned: true}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "BIGINT UNSIGNED", result)
}

func TestMySQL_CompileColumnType_BigIntegerAutoIncrement(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "id", Type: schema.TypeBigInteger, IsAutoIncrement: true}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "BIGINT AUTO_INCREMENT", result)
}

func TestMySQL_CompileColumnType_BigIntegerUnsignedAutoIncrement(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "id", Type: schema.TypeBigInteger, IsUnsigned: true, IsAutoIncrement: true}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "BIGINT UNSIGNED AUTO_INCREMENT", result)
}

func TestMySQL_CompileColumnType_Boolean(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "active", Type: schema.TypeBoolean}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "TINYINT(1)", result)
}

func TestMySQL_CompileColumnType_Timestamp(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "created_at", Type: schema.TypeTimestamp}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "TIMESTAMP", result)
}

func TestMySQL_CompileColumnType_Date(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "birth_date", Type: schema.TypeDate}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "DATE", result)
}

func TestMySQL_CompileColumnType_Decimal(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "price", Type: schema.TypeDecimal, Precision: 10, Scale: 2}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "DECIMAL(10, 2)", result)
}

func TestMySQL_CompileColumnType_DecimalDefaults(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "amount", Type: schema.TypeDecimal}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "DECIMAL(10, 0)", result)
}

func TestMySQL_CompileColumnType_Float(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "value", Type: schema.TypeFloat}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "DOUBLE", result)
}

func TestMySQL_CompileColumnType_UUID(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "external_id", Type: schema.TypeUUID}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "CHAR(36)", result)
}

func TestMySQL_CompileColumnType_JSON(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "metadata", Type: schema.TypeJSON}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "JSON", result)
}

func TestMySQL_CompileColumnType_Binary(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "data", Type: schema.TypeBinary}
	result, err := g.CompileColumnType(col)
	require.NoError(t, err)
	assert.Equal(t, "BLOB", result)
}

func TestMySQL_CompileColumnType_Unsupported(t *testing.T) {
	g := newMySQLGrammar()
	col := schema.ColumnDefinition{Name: "bad", Type: schema.ColumnType(999)}
	_, err := g.CompileColumnType(col)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnsupportedType))
	assert.Contains(t, err.Error(), `"bad"`)
}

// --- CompileCreate tests ---

func TestMySQL_CompileCreate_SimpleTable(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("users")
	bp.ID()
	bp.String("name", 255)

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, "CREATE TABLE `users`")
	assert.Contains(t, sql, "`id` BIGINT UNSIGNED AUTO_INCREMENT NOT NULL")
	assert.Contains(t, sql, "`name` VARCHAR(255) NOT NULL")
	assert.Contains(t, sql, "PRIMARY KEY (`id`)")
}

func TestMySQL_CompileCreate_NoColumns(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("empty")

	_, err := g.CompileCreate(bp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no columns defined")
}

func TestMySQL_CompileCreate_AllColumnTypes(t *testing.T) {
	g := newMySQLGrammar()
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
	assert.Contains(t, sql, "INT")
	assert.Contains(t, sql, "BIGINT")
	assert.Contains(t, sql, "TINYINT(1)")
	assert.Contains(t, sql, "TIMESTAMP")
	assert.Contains(t, sql, "DATE")
	assert.Contains(t, sql, "DECIMAL(8, 2)")
	assert.Contains(t, sql, "DOUBLE")
	assert.Contains(t, sql, "CHAR(36)")
	assert.Contains(t, sql, "JSON")
	assert.Contains(t, sql, "BLOB")
}

func TestMySQL_CompileCreate_NullableColumn(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("posts")
	bp.Text("body").Nullable()

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.NotContains(t, sql, "NOT NULL")
	assert.Contains(t, sql, "`body` TEXT")
}

func TestMySQL_CompileCreate_DefaultValue(t *testing.T) {
	g := newMySQLGrammar()
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

func TestMySQL_CompileCreate_UniqueColumn(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("users")
	bp.String("email", 255).Unique()

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, "UNIQUE")
}

func TestMySQL_CompileCreate_WithIndexes(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("users")
	bp.String("name", 255)
	bp.String("email", 255)
	bp.Index("name")
	bp.UniqueIndex("email")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, "UNIQUE KEY `uniq_users_email` (`email`)")
	assert.Contains(t, sql, "KEY `idx_users_name` (`name`)")
}

func TestMySQL_CompileCreate_WithForeignKey(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("posts")
	bp.BigInteger("user_id")
	bp.Foreign("user_id").References("id").On("users").OnDeleteAction("CASCADE").OnUpdateAction("SET NULL")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, "CONSTRAINT `fk_posts_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE SET NULL")
}

func TestMySQL_CompileCreate_FullTable(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("users")
	bp.ID()
	bp.String("name", 255)
	bp.String("email", 255).Unique()
	bp.Boolean("active").Default(true)
	bp.Timestamps()
	bp.Index("name")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, "CREATE TABLE `users`")
	assert.Contains(t, sql, "`id` BIGINT UNSIGNED AUTO_INCREMENT NOT NULL")
	assert.Contains(t, sql, "`name` VARCHAR(255) NOT NULL")
	assert.Contains(t, sql, "`email` VARCHAR(255) NOT NULL UNIQUE")
	assert.Contains(t, sql, "`active` TINYINT(1) NOT NULL DEFAULT TRUE")
	assert.Contains(t, sql, "`created_at` TIMESTAMP")
	assert.Contains(t, sql, "`updated_at` TIMESTAMP")
	assert.Contains(t, sql, "PRIMARY KEY (`id`)")
	assert.Contains(t, sql, "KEY `idx_users_name` (`name`)")
}

// --- CompileAlter tests ---

func TestMySQL_CompileAlter_AddColumn(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("users")
	bp.String("phone", 20).Nullable()

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, "ALTER TABLE `users` ADD COLUMN `phone` VARCHAR(20)", stmts[0])
}

func TestMySQL_CompileAlter_DropColumn(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("users")
	bp.DropColumn("age")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, "ALTER TABLE `users` DROP COLUMN `age`", stmts[0])
}

func TestMySQL_CompileAlter_RenameColumn(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("users")
	bp.RenameColumn("name", "full_name")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, "ALTER TABLE `users` RENAME COLUMN `name` TO `full_name`", stmts[0])
}

func TestMySQL_CompileAlter_DropIndex(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("users")
	bp.DropIndex("idx_users_email")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, "ALTER TABLE `users` DROP INDEX `idx_users_email`", stmts[0])
}

func TestMySQL_CompileAlter_DropForeign(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("posts")
	bp.DropForeign("fk_posts_user_id")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, "ALTER TABLE `posts` DROP FOREIGN KEY `fk_posts_user_id`", stmts[0])
}

func TestMySQL_CompileAlter_AddIndex(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("users")
	bp.Index("name")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, "CREATE INDEX `idx_users_name` ON `users` (`name`)", stmts[0])
}

func TestMySQL_CompileAlter_AddUniqueIndex(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("users")
	bp.UniqueIndex("email")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, "CREATE UNIQUE INDEX `uniq_users_email` ON `users` (`email`)", stmts[0])
}

func TestMySQL_CompileAlter_AddForeignKey(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("posts")
	bp.Foreign("user_id").References("id").On("users").OnDeleteAction("CASCADE")

	stmts, err := g.CompileAlter(bp)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Contains(t, stmts[0], "ALTER TABLE `posts` ADD CONSTRAINT `fk_posts_user_id`")
	assert.Contains(t, stmts[0], "FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE")
}

func TestMySQL_CompileAlter_MultipleOperations(t *testing.T) {
	g := newMySQLGrammar()
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

func TestMySQL_CompileDrop(t *testing.T) {
	g := newMySQLGrammar()
	assert.Equal(t, "DROP TABLE `users`", g.CompileDrop("users"))
}

func TestMySQL_CompileDropIfExists(t *testing.T) {
	g := newMySQLGrammar()
	assert.Equal(t, "DROP TABLE IF EXISTS `users`", g.CompileDropIfExists("users"))
}

// --- CompileRename test ---

func TestMySQL_CompileRename(t *testing.T) {
	g := newMySQLGrammar()
	assert.Equal(t, "RENAME TABLE `old_table` TO `new_table`", g.CompileRename("old_table", "new_table"))
}

// --- CompileHasTable test ---

func TestMySQL_CompileHasTable(t *testing.T) {
	g := newMySQLGrammar()
	sql := g.CompileHasTable("users")
	assert.Contains(t, sql, "information_schema.tables")
	assert.Contains(t, sql, "table_name = 'users'")
	assert.Contains(t, sql, "table_schema = DATABASE()")
}

// --- CompileHasColumn test ---

func TestMySQL_CompileHasColumn(t *testing.T) {
	g := newMySQLGrammar()
	sql := g.CompileHasColumn("users", "email")
	assert.Contains(t, sql, "information_schema.columns")
	assert.Contains(t, sql, "table_name = 'users'")
	assert.Contains(t, sql, "column_name = 'email'")
	assert.Contains(t, sql, "table_schema = DATABASE()")
}

// --- CompileDropAllTables test ---

func TestMySQL_CompileDropAllTables(t *testing.T) {
	g := newMySQLGrammar()
	sql := g.CompileDropAllTables()
	assert.Contains(t, sql, "SET FOREIGN_KEY_CHECKS = 0")
}

// --- Grammar interface compliance ---

func TestMySQLGrammar_ImplementsGrammar(t *testing.T) {
	var _ schema.Grammar = (*MySQLGrammar)(nil)
}

// --- Edge cases ---

func TestMySQL_CompileCreate_CompositeIndex(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("users")
	bp.String("first_name", 100)
	bp.String("last_name", 100)
	bp.Index("first_name", "last_name")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Contains(t, sql, "KEY `idx_users_first_name_last_name` (`first_name`, `last_name`)")
}

func TestMySQL_CompileCreate_DefaultStringEscapesSingleQuotes(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("test")
	bp.String("val", 100).Default("it's")

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Contains(t, sql, "DEFAULT 'it''s'")
}

func TestMySQL_CompileCreate_Timestamps(t *testing.T) {
	g := newMySQLGrammar()
	bp := schema.NewBlueprint("posts")
	bp.String("title", 255)
	bp.Timestamps()

	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)

	assert.Contains(t, sql, "`created_at` TIMESTAMP")
	assert.Contains(t, sql, "`updated_at` TIMESTAMP")
	// Timestamps should be nullable (no NOT NULL)
	parts := strings.Split(sql, ",")
	for _, part := range parts {
		if strings.Contains(part, "created_at") || strings.Contains(part, "updated_at") {
			assert.NotContains(t, part, "NOT NULL")
		}
	}
}
