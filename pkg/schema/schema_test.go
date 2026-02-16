package schema_test

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/andrianprasetya/go-migration/pkg/schema"
	"github.com/andrianprasetya/go-migration/pkg/schema/grammars"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)
	assert.NotNil(t, builder)
}

func TestBuilder_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(0, 0))

	err = builder.Create("users", func(bp *schema.Blueprint) {
		bp.ID()
		bp.String("name", 255)
		bp.String("email", 255).Unique()
		bp.Timestamps()
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuilder_Create_CompileError(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	// Empty blueprint should cause a compile error (no columns defined)
	err = builder.Create("empty_table", func(bp *schema.Blueprint) {
		// no columns
	})

	assert.Error(t, err)
}

func TestBuilder_Create_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	mock.ExpectExec("CREATE TABLE").WillReturnError(fmt.Errorf("connection lost"))

	err = builder.Create("users", func(bp *schema.Blueprint) {
		bp.ID()
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection lost")
}

func TestBuilder_Alter(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	mock.ExpectExec("ALTER TABLE").WillReturnResult(sqlmock.NewResult(0, 0))

	err = builder.Alter("users", func(bp *schema.Blueprint) {
		bp.String("phone", 20).Nullable()
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuilder_Alter_MultipleStatements(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	// Expect two ALTER TABLE statements: one for adding a column, one for dropping
	mock.ExpectExec("ALTER TABLE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("ALTER TABLE").WillReturnResult(sqlmock.NewResult(0, 0))

	err = builder.Alter("users", func(bp *schema.Blueprint) {
		bp.String("phone", 20).Nullable()
		bp.DropColumn("legacy_field")
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuilder_Alter_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	mock.ExpectExec("ALTER TABLE").WillReturnError(fmt.Errorf("table not found"))

	err = builder.Alter("nonexistent", func(bp *schema.Blueprint) {
		bp.String("col", 100)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "table not found")
}

func TestBuilder_Drop(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	mock.ExpectExec("DROP TABLE").WillReturnResult(sqlmock.NewResult(0, 0))

	err = builder.Drop("users")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuilder_Drop_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	mock.ExpectExec("DROP TABLE").WillReturnError(fmt.Errorf("permission denied"))

	err = builder.Drop("users")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestBuilder_DropIfExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	mock.ExpectExec("DROP TABLE IF EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))

	err = builder.DropIfExists("users")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuilder_Rename(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	mock.ExpectExec("ALTER TABLE").WillReturnResult(sqlmock.NewResult(0, 0))

	err = builder.Rename("old_table", "new_table")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuilder_HasTable_True(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(rows)

	exists, err := builder.HasTable("users")

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuilder_HasTable_False(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(rows)

	exists, err := builder.HasTable("nonexistent")

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuilder_HasTable_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	mock.ExpectQuery("SELECT COUNT").WillReturnError(fmt.Errorf("query failed"))

	exists, err := builder.HasTable("users")

	assert.Error(t, err)
	assert.False(t, exists)
}

func TestBuilder_HasColumn_True(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(rows)

	exists, err := builder.HasColumn("users", "email")

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuilder_HasColumn_False(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(rows)

	exists, err := builder.HasColumn("users", "nonexistent")

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuilder_HasColumn_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()
	builder := schema.NewBuilder(db, grammar)

	mock.ExpectQuery("SELECT COUNT").WillReturnError(fmt.Errorf("query failed"))

	exists, err := builder.HasColumn("users", "email")

	assert.Error(t, err)
	assert.False(t, exists)
}

func TestBuilder_WithTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewPostgresGrammar()

	// Begin a transaction and use it as the executor
	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)

	builder := schema.NewBuilder(tx, grammar)

	mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(0, 0))

	err = builder.Create("posts", func(bp *schema.Blueprint) {
		bp.ID()
		bp.String("title", 255)
		bp.Text("body").Nullable()
	})

	assert.NoError(t, err)

	mock.ExpectCommit()
	err = tx.Commit()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuilder_WithMySQLGrammar(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewMySQLGrammar()
	builder := schema.NewBuilder(db, grammar)

	mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(0, 0))

	err = builder.Create("products", func(bp *schema.Blueprint) {
		bp.ID()
		bp.String("name", 100)
		bp.Decimal("price", 10, 2)
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuilder_WithSQLiteGrammar(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grammar := grammars.NewSQLiteGrammar()
	builder := schema.NewBuilder(db, grammar)

	mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(0, 0))

	err = builder.Create("settings", func(bp *schema.Blueprint) {
		bp.ID()
		bp.String("key", 100).Unique()
		bp.Text("value")
	})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
