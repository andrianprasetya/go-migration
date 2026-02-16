package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBlueprint(t *testing.T) {
	bp := NewBlueprint("users")
	assert.Equal(t, "users", bp.Table())
	assert.Empty(t, bp.Columns())
	assert.Empty(t, bp.Indexes())
	assert.Empty(t, bp.ForeignKeys())
	assert.Empty(t, bp.Commands())
}

// --- Column type methods ---

func TestBlueprint_ID(t *testing.T) {
	bp := NewBlueprint("users")
	col := bp.ID()

	assert.Equal(t, "id", col.Name)
	assert.Equal(t, TypeBigInteger, col.Type)
	assert.True(t, col.IsPrimary)
	assert.True(t, col.IsAutoIncrement)
	assert.True(t, col.IsUnsigned)
}

func TestBlueprint_String(t *testing.T) {
	bp := NewBlueprint("users")
	col := bp.String("name", 255)

	assert.Equal(t, "name", col.Name)
	assert.Equal(t, TypeString, col.Type)
	assert.Equal(t, 255, col.Length)
}

func TestBlueprint_Text(t *testing.T) {
	bp := NewBlueprint("posts")
	col := bp.Text("body")

	assert.Equal(t, "body", col.Name)
	assert.Equal(t, TypeText, col.Type)
}

func TestBlueprint_Integer(t *testing.T) {
	bp := NewBlueprint("users")
	col := bp.Integer("age")

	assert.Equal(t, "age", col.Name)
	assert.Equal(t, TypeInteger, col.Type)
}

func TestBlueprint_BigInteger(t *testing.T) {
	bp := NewBlueprint("orders")
	col := bp.BigInteger("total")

	assert.Equal(t, "total", col.Name)
	assert.Equal(t, TypeBigInteger, col.Type)
}

func TestBlueprint_Boolean(t *testing.T) {
	bp := NewBlueprint("users")
	col := bp.Boolean("active")

	assert.Equal(t, "active", col.Name)
	assert.Equal(t, TypeBoolean, col.Type)
}

func TestBlueprint_Timestamp(t *testing.T) {
	bp := NewBlueprint("events")
	col := bp.Timestamp("occurred_at")

	assert.Equal(t, "occurred_at", col.Name)
	assert.Equal(t, TypeTimestamp, col.Type)
}

func TestBlueprint_Date(t *testing.T) {
	bp := NewBlueprint("events")
	col := bp.Date("event_date")

	assert.Equal(t, "event_date", col.Name)
	assert.Equal(t, TypeDate, col.Type)
}

func TestBlueprint_Decimal(t *testing.T) {
	bp := NewBlueprint("products")
	col := bp.Decimal("price", 10, 2)

	assert.Equal(t, "price", col.Name)
	assert.Equal(t, TypeDecimal, col.Type)
	assert.Equal(t, 10, col.Precision)
	assert.Equal(t, 2, col.Scale)
}

func TestBlueprint_Float(t *testing.T) {
	bp := NewBlueprint("measurements")
	col := bp.Float("value")

	assert.Equal(t, "value", col.Name)
	assert.Equal(t, TypeFloat, col.Type)
}

func TestBlueprint_UUID(t *testing.T) {
	bp := NewBlueprint("users")
	col := bp.UUID("external_id")

	assert.Equal(t, "external_id", col.Name)
	assert.Equal(t, TypeUUID, col.Type)
}

func TestBlueprint_JSON(t *testing.T) {
	bp := NewBlueprint("settings")
	col := bp.JSON("metadata")

	assert.Equal(t, "metadata", col.Name)
	assert.Equal(t, TypeJSON, col.Type)
}

func TestBlueprint_Binary(t *testing.T) {
	bp := NewBlueprint("files")
	col := bp.Binary("data")

	assert.Equal(t, "data", col.Name)
	assert.Equal(t, TypeBinary, col.Type)
}

// --- Convenience methods ---

func TestBlueprint_Timestamps(t *testing.T) {
	bp := NewBlueprint("users")
	bp.Timestamps()

	cols := bp.Columns()
	require.Len(t, cols, 2)

	assert.Equal(t, "created_at", cols[0].Name)
	assert.Equal(t, TypeTimestamp, cols[0].Type)
	assert.True(t, cols[0].IsNullable)

	assert.Equal(t, "updated_at", cols[1].Name)
	assert.Equal(t, TypeTimestamp, cols[1].Type)
	assert.True(t, cols[1].IsNullable)
}

func TestBlueprint_SoftDeletes(t *testing.T) {
	bp := NewBlueprint("users")
	bp.SoftDeletes()

	cols := bp.Columns()
	require.Len(t, cols, 1)

	assert.Equal(t, "deleted_at", cols[0].Name)
	assert.Equal(t, TypeTimestamp, cols[0].Type)
	assert.True(t, cols[0].IsNullable)
}

// --- Chaining modifiers through Blueprint ---

func TestBlueprint_ColumnModifierChaining(t *testing.T) {
	bp := NewBlueprint("users")
	bp.String("email", 255).Nullable().Unique().Default("")

	cols := bp.Columns()
	require.Len(t, cols, 1)
	assert.True(t, cols[0].IsNullable)
	assert.True(t, cols[0].IsUnique)
	assert.Equal(t, "", cols[0].DefaultValue)
}

// --- Index methods ---

func TestBlueprint_Index(t *testing.T) {
	bp := NewBlueprint("users")
	idx := bp.Index("email")

	assert.Equal(t, "idx_users_email", idx.Name)
	assert.Equal(t, []string{"email"}, idx.Columns)
	assert.False(t, idx.Unique)
}

func TestBlueprint_Index_Composite(t *testing.T) {
	bp := NewBlueprint("users")
	idx := bp.Index("first_name", "last_name")

	assert.Equal(t, "idx_users_first_name_last_name", idx.Name)
	assert.Equal(t, []string{"first_name", "last_name"}, idx.Columns)
	assert.False(t, idx.Unique)
}

func TestBlueprint_UniqueIndex(t *testing.T) {
	bp := NewBlueprint("users")
	idx := bp.UniqueIndex("email")

	assert.Equal(t, "uniq_users_email", idx.Name)
	assert.Equal(t, []string{"email"}, idx.Columns)
	assert.True(t, idx.Unique)
}

// --- Foreign key methods ---

func TestBlueprint_Foreign(t *testing.T) {
	bp := NewBlueprint("posts")
	fk := bp.Foreign("user_id").References("id").On("users").OnDeleteAction("CASCADE").OnUpdateAction("SET NULL")

	assert.Equal(t, "user_id", fk.Column)
	assert.Equal(t, "id", fk.RefColumn)
	assert.Equal(t, "users", fk.RefTable)
	assert.Equal(t, "CASCADE", fk.OnDelete)
	assert.Equal(t, "SET NULL", fk.OnUpdate)
	assert.Equal(t, "fk_posts_user_id", fk.Name)
}

func TestBlueprint_Foreign_DefaultName(t *testing.T) {
	bp := NewBlueprint("comments")
	fk := bp.Foreign("post_id")

	assert.Equal(t, "fk_comments_post_id", fk.Name)
}

// --- Alter methods ---

func TestBlueprint_DropColumn(t *testing.T) {
	bp := NewBlueprint("users")
	bp.DropColumn("age")

	cmds := bp.Commands()
	require.Len(t, cmds, 1)
	assert.Equal(t, CommandDropColumn, cmds[0].Type)
	assert.Equal(t, "age", cmds[0].Name)
}

func TestBlueprint_RenameColumn(t *testing.T) {
	bp := NewBlueprint("users")
	bp.RenameColumn("name", "full_name")

	cmds := bp.Commands()
	require.Len(t, cmds, 1)
	assert.Equal(t, CommandRenameColumn, cmds[0].Type)
	assert.Equal(t, "name", cmds[0].Name)
	assert.Equal(t, "full_name", cmds[0].To)
}

func TestBlueprint_DropIndex(t *testing.T) {
	bp := NewBlueprint("users")
	bp.DropIndex("idx_users_email")

	cmds := bp.Commands()
	require.Len(t, cmds, 1)
	assert.Equal(t, CommandDropIndex, cmds[0].Type)
	assert.Equal(t, "idx_users_email", cmds[0].Name)
}

func TestBlueprint_DropForeign(t *testing.T) {
	bp := NewBlueprint("posts")
	bp.DropForeign("fk_posts_user_id")

	cmds := bp.Commands()
	require.Len(t, cmds, 1)
	assert.Equal(t, CommandDropForeign, cmds[0].Type)
	assert.Equal(t, "fk_posts_user_id", cmds[0].Name)
}

// --- Accessor methods return copies ---

func TestBlueprint_Columns_ReturnsCopy(t *testing.T) {
	bp := NewBlueprint("users")
	bp.String("name", 100)

	cols1 := bp.Columns()
	cols1[0].Name = "mutated"

	cols2 := bp.Columns()
	assert.Equal(t, "name", cols2[0].Name, "mutating returned slice should not affect blueprint")
}

// --- Full table definition ---

func TestBlueprint_FullTableDefinition(t *testing.T) {
	bp := NewBlueprint("users")
	bp.ID()
	bp.String("name", 255)
	bp.String("email", 255).Unique()
	bp.Boolean("active").Default(true)
	bp.Timestamps()
	bp.SoftDeletes()
	bp.Index("name")
	bp.Foreign("team_id").References("id").On("teams").OnDeleteAction("CASCADE")

	assert.Equal(t, 7, len(bp.Columns())) // id, name, email, active, created_at, updated_at, deleted_at
	assert.Equal(t, 1, len(bp.Indexes()))
	assert.Equal(t, 1, len(bp.ForeignKeys()))
}
