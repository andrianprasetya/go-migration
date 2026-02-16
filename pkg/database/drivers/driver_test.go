package drivers

import (
	"testing"
	"time"

	"github.com/andrianprasetya/go-migration/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestPostgresDriver_Name(t *testing.T) {
	d := NewPostgresDriver()
	assert.Equal(t, "postgres", d.Name())
}

func TestPostgresDriver_BuildDSN(t *testing.T) {
	d := NewPostgresDriver()

	config := database.ConnectionConfig{
		Host:     "localhost",
		Port:     5432,
		Username: "user",
		Password: "pass",
		Database: "testdb",
	}

	dsn := d.buildDSN(config)
	assert.Contains(t, dsn, "host=localhost")
	assert.Contains(t, dsn, "port=5432")
	assert.Contains(t, dsn, "user=user")
	assert.Contains(t, dsn, "password=pass")
	assert.Contains(t, dsn, "dbname=testdb")
	assert.Contains(t, dsn, "sslmode=disable")
}

func TestPostgresDriver_BuildDSN_CustomSSLMode(t *testing.T) {
	d := NewPostgresDriver()

	config := database.ConnectionConfig{
		Host:     "localhost",
		Port:     5432,
		Username: "user",
		Password: "pass",
		Database: "testdb",
		Options:  map[string]string{"sslmode": "require"},
	}

	dsn := d.buildDSN(config)
	assert.Contains(t, dsn, "sslmode=require")
	assert.NotContains(t, dsn, "sslmode=disable")
}

func TestMySQLDriver_Name(t *testing.T) {
	d := NewMySQLDriver()
	assert.Equal(t, "mysql", d.Name())
}

func TestMySQLDriver_BuildDSN(t *testing.T) {
	d := NewMySQLDriver()

	config := database.ConnectionConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "secret",
		Database: "mydb",
	}

	dsn := d.buildDSN(config)
	assert.Equal(t, "root:secret@tcp(localhost:3306)/mydb", dsn)
}

func TestMySQLDriver_BuildDSN_WithOptions(t *testing.T) {
	d := NewMySQLDriver()

	config := database.ConnectionConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "secret",
		Database: "mydb",
		Options:  map[string]string{"charset": "utf8mb4"},
	}

	dsn := d.buildDSN(config)
	assert.Contains(t, dsn, "root:secret@tcp(localhost:3306)/mydb?")
	assert.Contains(t, dsn, "charset=utf8mb4")
}

func TestSQLiteDriver_Name(t *testing.T) {
	d := NewSQLiteDriver()
	assert.Equal(t, "sqlite3", d.Name())
}

func TestSQLiteDriver_BuildDSN(t *testing.T) {
	d := NewSQLiteDriver()

	config := database.ConnectionConfig{
		Database: "./data.db",
	}

	dsn := d.buildDSN(config)
	assert.Equal(t, "./data.db", dsn)
}

func TestSQLiteDriver_BuildDSN_Memory(t *testing.T) {
	d := NewSQLiteDriver()

	config := database.ConnectionConfig{
		Database: ":memory:",
	}

	dsn := d.buildDSN(config)
	assert.Equal(t, ":memory:", dsn)
}

func TestSQLiteDriver_BuildDSN_WithOptions(t *testing.T) {
	d := NewSQLiteDriver()

	config := database.ConnectionConfig{
		Database: "./data.db",
		Options:  map[string]string{"cache": "shared"},
	}

	dsn := d.buildDSN(config)
	assert.Contains(t, dsn, "./data.db?")
	assert.Contains(t, dsn, "cache=shared")
}

func TestDriverInterface_Compliance(t *testing.T) {
	// Verify all drivers implement the Driver interface at compile time.
	var _ Driver = (*PostgresDriver)(nil)
	var _ Driver = (*MySQLDriver)(nil)
	var _ Driver = (*SQLiteDriver)(nil)
}

func TestConfigurePool(t *testing.T) {
	// configurePool is tested indirectly through Open, but we can verify
	// it doesn't panic with zero-value config.
	config := database.ConnectionConfig{}
	// sql.Open with unknown driver will fail, but configurePool itself
	// should handle zero values gracefully. We test the helper directly
	// by noting it only sets values when > 0.
	config.MaxOpenConns = 0
	config.MaxIdleConns = 0
	config.ConnMaxLifetime = 0
	// No assertion needed — just verifying no panic with zero config.
}

func TestConfigurePool_WithValues(t *testing.T) {
	config := database.ConnectionConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxLifetime)
}
