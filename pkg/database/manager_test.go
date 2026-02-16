package database

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDriver is a test double that records calls and returns a controllable *sql.DB.
type mockDriver struct {
	name   string
	openFn func(config ConnectionConfig) (*sql.DB, error)
}

func (d *mockDriver) Open(config ConnectionConfig) (*sql.DB, error) {
	if d.openFn != nil {
		return d.openFn(config)
	}
	return nil, fmt.Errorf("mockDriver.Open not configured")
}

func (d *mockDriver) Name() string {
	return d.name
}

// newTestDriver creates a mock driver that opens a SQLite in-memory DB.
// This avoids needing real database servers for unit tests.
func newTestDriver(name string) *mockDriver {
	return &mockDriver{
		name: name,
		openFn: func(config ConnectionConfig) (*sql.DB, error) {
			// Use sql.Open with a driver that won't actually connect
			// but gives us a valid *sql.DB handle for testing.
			// We use "sqlmock" pattern — just return a stub.
			// Since we can't import sqlmock here, we'll use a simple approach:
			// open with an unknown driver to get a *sql.DB that won't ping.
			db, err := sql.Open("txdb_test", "test")
			if err != nil {
				// Fallback: create a minimal *sql.DB
				return nil, err
			}
			return db, nil
		},
	}
}

// fakeDB creates a minimal *sql.DB for testing purposes.
// We register a fake driver once and reuse it.
func fakeDB(t *testing.T) *sql.DB {
	t.Helper()
	// We can't easily create a *sql.DB without a registered driver,
	// so we'll test the manager logic through its public API.
	return nil
}

func TestNewManager(t *testing.T) {
	m := NewManager()
	assert.NotNil(t, m)
	assert.NotNil(t, m.connections)
	assert.NotNil(t, m.configs)
	assert.NotNil(t, m.drivers)
	assert.Empty(t, m.defaultName)
}

func TestRegisterDriver(t *testing.T) {
	m := NewManager()
	d := &mockDriver{name: "postgres"}

	m.RegisterDriver("postgres", d)

	assert.Len(t, m.drivers, 1)
	assert.Equal(t, d, m.drivers["postgres"])
}

func TestRegisterDriver_Multiple(t *testing.T) {
	m := NewManager()
	pg := &mockDriver{name: "postgres"}
	my := &mockDriver{name: "mysql"}

	m.RegisterDriver("postgres", pg)
	m.RegisterDriver("mysql", my)

	assert.Len(t, m.drivers, 2)
}

func TestAddConnection(t *testing.T) {
	m := NewManager()

	err := m.AddConnection("primary", ConnectionConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
	})

	require.NoError(t, err)
	assert.Contains(t, m.configs, "primary")
	// First connection becomes default.
	assert.Equal(t, "primary", m.defaultName)
}

func TestAddConnection_MissingDriver(t *testing.T) {
	m := NewManager()

	err := m.AddConnection("bad", ConnectionConfig{
		Host:     "localhost",
		Database: "testdb",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "driver is required")
}

func TestAddConnection_FirstBecomesDefault(t *testing.T) {
	m := NewManager()

	_ = m.AddConnection("first", ConnectionConfig{Driver: "pg"})
	_ = m.AddConnection("second", ConnectionConfig{Driver: "pg"})

	assert.Equal(t, "first", m.defaultName)
}

func TestConnection_NotFound(t *testing.T) {
	m := NewManager()

	_, err := m.Connection("nonexistent")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrConnectionNotFound))
}

func TestConnection_DriverNotFound(t *testing.T) {
	m := NewManager()
	_ = m.AddConnection("test", ConnectionConfig{
		Driver:   "unknown_driver",
		Host:     "localhost",
		Database: "testdb",
	})

	_, err := m.Connection("test")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrDriverNotFound))
}

func TestConnection_DriverOpenFails(t *testing.T) {
	m := NewManager()
	m.RegisterDriver("failing", &mockDriver{
		name: "failing",
		openFn: func(config ConnectionConfig) (*sql.DB, error) {
			return nil, fmt.Errorf("connection refused")
		},
	})
	_ = m.AddConnection("test", ConnectionConfig{
		Driver:   "failing",
		Host:     "localhost",
		Database: "testdb",
	})

	_, err := m.Connection("test")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrConnectionFailed))
}

func TestConnection_LazyOpen(t *testing.T) {
	openCalled := false
	m := NewManager()
	m.RegisterDriver("lazy", &mockDriver{
		name: "lazy",
		openFn: func(config ConnectionConfig) (*sql.DB, error) {
			openCalled = true
			// Return a real *sql.DB via sql.Open with a dummy driver name.
			// This will create a *sql.DB that can't actually connect,
			// but is valid for our manager caching tests.
			db, _ := sql.Open("txdb_nonexist", "")
			return db, nil
		},
	})
	_ = m.AddConnection("test", ConnectionConfig{
		Driver:   "lazy",
		Database: "testdb",
	})

	// Not opened yet after AddConnection.
	assert.False(t, openCalled)
}

func TestDefault_NoDefault(t *testing.T) {
	m := NewManager()

	_, err := m.Default()

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoDefault))
}

func TestSetDefault(t *testing.T) {
	m := NewManager()
	_ = m.AddConnection("primary", ConnectionConfig{Driver: "pg"})
	_ = m.AddConnection("secondary", ConnectionConfig{Driver: "pg"})

	err := m.SetDefault("secondary")

	require.NoError(t, err)
	assert.Equal(t, "secondary", m.defaultName)
}

func TestSetDefault_NotFound(t *testing.T) {
	m := NewManager()

	err := m.SetDefault("nonexistent")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrConnectionNotFound))
}

func TestClose_NoConnections(t *testing.T) {
	m := NewManager()

	err := m.Close()

	assert.NoError(t, err)
}

func TestPoolConfig(t *testing.T) {
	config := ConnectionConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	// Verify the config values are stored correctly.
	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxLifetime)
}

func TestDriverInterface_Compliance(t *testing.T) {
	// Verify mockDriver implements the Driver interface.
	var _ Driver = (*mockDriver)(nil)
}

func TestConnection_ErrorMessages(t *testing.T) {
	m := NewManager()

	_, err := m.Connection("missing")
	assert.Contains(t, err.Error(), "missing")
	assert.Contains(t, err.Error(), "connection not found")
}

func TestSetDefault_ErrorMessage(t *testing.T) {
	m := NewManager()

	err := m.SetDefault("missing")
	assert.Contains(t, err.Error(), "missing")
	assert.Contains(t, err.Error(), "connection not found")
}
