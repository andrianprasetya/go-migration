package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColumnType_String(t *testing.T) {
	tests := []struct {
		ct   ColumnType
		want string
	}{
		{TypeString, "string"},
		{TypeText, "text"},
		{TypeInteger, "integer"},
		{TypeBigInteger, "bigInteger"},
		{TypeBoolean, "boolean"},
		{TypeTimestamp, "timestamp"},
		{TypeDate, "date"},
		{TypeDecimal, "decimal"},
		{TypeFloat, "float"},
		{TypeUUID, "uuid"},
		{TypeJSON, "json"},
		{TypeBinary, "binary"},
		{ColumnType(999), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ct.String())
		})
	}
}

func TestColumnDefinition_Chainable(t *testing.T) {
	col := &ColumnDefinition{Name: "email", Type: TypeString}

	result := col.Nullable().Default("test@example.com").Unique()

	assert.Same(t, col, result, "modifiers should return the same pointer for chaining")
	assert.True(t, col.IsNullable)
	assert.Equal(t, "test@example.com", col.DefaultValue)
	assert.True(t, col.IsUnique)
}

func TestColumnDefinition_AllModifiers(t *testing.T) {
	col := &ColumnDefinition{Name: "id", Type: TypeBigInteger}

	col.Primary().Unsigned().AutoIncrement().Default(0).Nullable()

	assert.True(t, col.IsPrimary)
	assert.True(t, col.IsUnsigned)
	assert.True(t, col.IsAutoIncrement)
	assert.Equal(t, 0, col.DefaultValue)
	assert.True(t, col.IsNullable)
}
