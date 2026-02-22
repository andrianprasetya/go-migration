package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestLoadFromYAML(t *testing.T) {
	content := `
default: primary
migration_table: schema_migrations
connections:
  primary:
    driver: postgres
    host: localhost
    port: 5432
    database: testdb
    username: user
    password: pass
  secondary:
    driver: mysql
    host: 127.0.0.1
    port: 3306
    database: otherdb
    username: root
    password: secret
`
	path := writeTestFile(t, "config.yaml", content)

	cfg, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, "primary", cfg.DefaultConn)
	assert.Equal(t, "schema_migrations", cfg.MigrationTable)
	assert.Len(t, cfg.Connections, 2)

	primary := cfg.Connections["primary"]
	assert.Equal(t, "postgres", primary.Driver)
	assert.Equal(t, "localhost", primary.Host)
	assert.Equal(t, 5432, primary.Port)
	assert.Equal(t, "testdb", primary.Database)

	secondary := cfg.Connections["secondary"]
	assert.Equal(t, "mysql", secondary.Driver)
	assert.Equal(t, "127.0.0.1", secondary.Host)
	assert.Equal(t, 3306, secondary.Port)
}

func TestLoadFromYML(t *testing.T) {
	content := `
connections:
  main:
    driver: sqlite
    host: ":memory:"
    database: test.db
`
	path := writeTestFile(t, "config.yml", content)

	cfg, err := Load(path)
	require.NoError(t, err)
	assert.Len(t, cfg.Connections, 1)
	assert.Equal(t, "sqlite", cfg.Connections["main"].Driver)
}

func TestLoadFromJSON(t *testing.T) {
	cfgData := Config{
		DefaultConn:    "main",
		MigrationTable: "my_migrations",
		Connections: map[string]ConnectionConfig{
			"main": {
				Driver:   "postgres",
				Host:     "db.example.com",
				Port:     5432,
				Database: "production",
				Username: "admin",
				Password: "secret",
			},
		},
	}

	data, err := json.MarshalIndent(cfgData, "", "  ")
	require.NoError(t, err)

	path := writeTestFile(t, "config.json", string(data))

	cfg, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, "main", cfg.DefaultConn)
	assert.Equal(t, "my_migrations", cfg.MigrationTable)
	assert.Equal(t, "postgres", cfg.Connections["main"].Driver)
	assert.Equal(t, "db.example.com", cfg.Connections["main"].Host)
}

func TestLoadUnsupportedFormat(t *testing.T) {
	path := writeTestFile(t, "config.toml", "key = value")

	_, err := Load(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported config file format")
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadFromEnv(t *testing.T) {
	envVars := map[string]string{
		"GOMIGRATE_DEFAULT_CONNECTION": "primary",
		"GOMIGRATE_DB_DRIVER":          "postgres",
		"GOMIGRATE_DB_HOST":            "localhost",
		"GOMIGRATE_DB_PORT":            "5432",
		"GOMIGRATE_DB_DATABASE":        "testdb",
		"GOMIGRATE_DB_USERNAME":        "user",
		"GOMIGRATE_DB_PASSWORD":        "pass",
		"GOMIGRATE_MIGRATION_TABLE":    "schema_migrations",
		"GOMIGRATE_MIGRATION_DIR":      "db/migrations",
		"GOMIGRATE_SEEDER_DIR":         "db/seeders",
		"GOMIGRATE_LOG_LEVEL":          "debug",
		"GOMIGRATE_LOG_OUTPUT":         "file",
	}
	setEnvVars(t, envVars)

	cfg, err := LoadFromEnv()
	require.NoError(t, err)

	assert.Equal(t, "primary", cfg.DefaultConn)
	assert.Equal(t, "schema_migrations", cfg.MigrationTable)
	assert.Equal(t, "db/migrations", cfg.MigrationDir)
	assert.Equal(t, "db/seeders", cfg.SeederDir)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "file", cfg.LogOutput)

	conn := cfg.Connections["primary"]
	assert.Equal(t, "postgres", conn.Driver)
	assert.Equal(t, "localhost", conn.Host)
	assert.Equal(t, 5432, conn.Port)
	assert.Equal(t, "testdb", conn.Database)
	assert.Equal(t, "user", conn.Username)
	assert.Equal(t, "pass", conn.Password)
}

func TestLoadFromEnvInvalidPort(t *testing.T) {
	setEnvVars(t, map[string]string{
		"GOMIGRATE_DB_DRIVER":   "postgres",
		"GOMIGRATE_DB_HOST":     "localhost",
		"GOMIGRATE_DB_PORT":     "not_a_number",
		"GOMIGRATE_DB_DATABASE": "testdb",
	})

	_, err := LoadFromEnv()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid GOMIGRATE_DB_PORT")
}

func TestLoadFromEnvNoDBVars(t *testing.T) {
	// No GOMIGRATE_ env vars set — LoadFromEnv should still work with defaults

	cfg, err := LoadFromEnv()
	require.NoError(t, err)

	// Should have defaults applied
	assert.Equal(t, "migrations", cfg.MigrationTable)
	assert.Equal(t, "migrations", cfg.MigrationDir)
	assert.Equal(t, "seeders", cfg.SeederDir)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "console", cfg.LogOutput)
}

func TestValidateSuccess(t *testing.T) {
	cfg := &Config{
		Connections: map[string]ConnectionConfig{
			"default": {
				Driver:   "postgres",
				Host:     "localhost",
				Database: "testdb",
			},
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidateNoConnections(t *testing.T) {
	cfg := &Config{
		Connections: map[string]ConnectionConfig{},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigValidation)
	assert.Contains(t, err.Error(), "at least one connection must be defined")
}

func TestValidateMissingFields(t *testing.T) {
	cfg := &Config{
		Connections: map[string]ConnectionConfig{
			"broken": {
				// Missing driver, host, database
				Username: "user",
			},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigValidation)
	assert.Contains(t, err.Error(), "connections.broken.driver is required")
	assert.Contains(t, err.Error(), "connections.broken.host is required")
	assert.Contains(t, err.Error(), "connections.broken.database is required")
}

func TestValidatePartialMissingFields(t *testing.T) {
	cfg := &Config{
		Connections: map[string]ConnectionConfig{
			"partial": {
				Driver: "mysql",
				// Missing host and database
			},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigValidation)
	assert.Contains(t, err.Error(), "connections.partial.host is required")
	assert.Contains(t, err.Error(), "connections.partial.database is required")
	assert.NotContains(t, err.Error(), "connections.partial.driver")
}

func TestApplyDefaults(t *testing.T) {
	cfg := &Config{}
	cfg.ApplyDefaults()

	assert.Equal(t, "migrations", cfg.MigrationTable)
	assert.Equal(t, "migrations", cfg.MigrationDir)
	assert.Equal(t, "seeders", cfg.SeederDir)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "console", cfg.LogOutput)
	assert.NotNil(t, cfg.Connections)
}

func TestApplyDefaultsPreservesExisting(t *testing.T) {
	cfg := &Config{
		MigrationTable: "custom_migrations",
		MigrationDir:   "db/migrate",
		SeederDir:      "db/seeds",
		LogLevel:       "debug",
		LogOutput:      "file",
	}
	cfg.ApplyDefaults()

	assert.Equal(t, "custom_migrations", cfg.MigrationTable)
	assert.Equal(t, "db/migrate", cfg.MigrationDir)
	assert.Equal(t, "db/seeds", cfg.SeederDir)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "file", cfg.LogOutput)
}

func TestConfigYAMLRoundTrip(t *testing.T) {
	original := &Config{
		DefaultConn:    "primary",
		MigrationTable: "migrations",
		MigrationDir:   "db/migrations",
		SeederDir:      "db/seeders",
		LogLevel:       "info",
		LogOutput:      "console",
		Connections: map[string]ConnectionConfig{
			"primary": {
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "admin",
				Password: "secret",
			},
		},
	}

	data, err := yaml.Marshal(original)
	require.NoError(t, err)

	path := writeTestFile(t, "roundtrip.yaml", string(data))
	loaded, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, original.DefaultConn, loaded.DefaultConn)
	assert.Equal(t, original.MigrationTable, loaded.MigrationTable)
	assert.Equal(t, original.Connections["primary"].Driver, loaded.Connections["primary"].Driver)
	assert.Equal(t, original.Connections["primary"].Host, loaded.Connections["primary"].Host)
	assert.Equal(t, original.Connections["primary"].Port, loaded.Connections["primary"].Port)
	assert.Equal(t, original.Connections["primary"].Database, loaded.Connections["primary"].Database)
}

func TestConfigJSONRoundTrip(t *testing.T) {
	original := &Config{
		DefaultConn:    "main",
		MigrationTable: "schema_migrations",
		Connections: map[string]ConnectionConfig{
			"main": {
				Driver:   "mysql",
				Host:     "db.example.com",
				Port:     3306,
				Database: "production",
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	path := writeTestFile(t, "roundtrip.json", string(data))
	loaded, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, original.DefaultConn, loaded.DefaultConn)
	assert.Equal(t, original.MigrationTable, loaded.MigrationTable)
	assert.Equal(t, original.Connections["main"].Driver, loaded.Connections["main"].Driver)
	assert.Equal(t, original.Connections["main"].Host, loaded.Connections["main"].Host)
}

// --- helpers ---

func writeTestFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}

func setEnvVars(t *testing.T, vars map[string]string) {
	t.Helper()
	for k, v := range vars {
		t.Setenv(k, v)
	}
}

func TestValidateInvalidDriver(t *testing.T) {
	cfg := &Config{
		Connections: map[string]ConnectionConfig{
			"default": {
				Driver:   "oracle",
				Host:     "localhost",
				Database: "testdb",
			},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigValidation)
	assert.Contains(t, err.Error(), "connections.default.driver must be one of: postgres, mysql, sqlite")
}

func TestValidateNegativePort(t *testing.T) {
	cfg := &Config{
		Connections: map[string]ConnectionConfig{
			"default": {
				Driver:   "postgres",
				Host:     "localhost",
				Port:     -1,
				Database: "testdb",
			},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigValidation)
	assert.Contains(t, err.Error(), "connections.default.port must be a positive integer")
}

func TestValidateZeroPortIsValid(t *testing.T) {
	cfg := &Config{
		Connections: map[string]ConnectionConfig{
			"default": {
				Driver:   "postgres",
				Host:     "localhost",
				Port:     0,
				Database: "testdb",
			},
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidateInvalidLogLevel(t *testing.T) {
	cfg := &Config{
		LogLevel: "verbose",
		Connections: map[string]ConnectionConfig{
			"default": {
				Driver:   "postgres",
				Host:     "localhost",
				Database: "testdb",
			},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigValidation)
	assert.Contains(t, err.Error(), "log_level must be one of: debug, info, warn, error")
}

func TestValidateInvalidLogOutput(t *testing.T) {
	cfg := &Config{
		LogOutput: "syslog",
		Connections: map[string]ConnectionConfig{
			"default": {
				Driver:   "postgres",
				Host:     "localhost",
				Database: "testdb",
			},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigValidation)
	assert.Contains(t, err.Error(), "log_output must be one of: console, file, both")
}

func TestValidateNegativeMaxOpenConns(t *testing.T) {
	cfg := &Config{
		Connections: map[string]ConnectionConfig{
			"default": {
				Driver:       "postgres",
				Host:         "localhost",
				Database:     "testdb",
				MaxOpenConns: -5,
			},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigValidation)
	assert.Contains(t, err.Error(), "connections.default.max_open_conns must be non-negative")
}

func TestValidateNegativeMaxIdleConns(t *testing.T) {
	cfg := &Config{
		Connections: map[string]ConnectionConfig{
			"default": {
				Driver:       "postgres",
				Host:         "localhost",
				Database:     "testdb",
				MaxIdleConns: -3,
			},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigValidation)
	assert.Contains(t, err.Error(), "connections.default.max_idle_conns must be non-negative")
}

func TestValidateMultipleViolations(t *testing.T) {
	cfg := &Config{
		LogLevel:  "verbose",
		LogOutput: "syslog",
		Connections: map[string]ConnectionConfig{
			"default": {
				Driver:       "oracle",
				Host:         "localhost",
				Database:     "testdb",
				Port:         -1,
				MaxOpenConns: -5,
				MaxIdleConns: -3,
			},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConfigValidation)
	errMsg := err.Error()
	assert.Contains(t, errMsg, "driver must be one of")
	assert.Contains(t, errMsg, "port must be a positive integer")
	assert.Contains(t, errMsg, "max_open_conns must be non-negative")
	assert.Contains(t, errMsg, "max_idle_conns must be non-negative")
	assert.Contains(t, errMsg, "log_level must be one of")
	assert.Contains(t, errMsg, "log_output must be one of")
}

func TestValidateValidLogLevels(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "error"} {
		cfg := &Config{
			LogLevel: level,
			Connections: map[string]ConnectionConfig{
				"default": {
					Driver:   "postgres",
					Host:     "localhost",
					Database: "testdb",
				},
			},
		}
		err := cfg.Validate()
		assert.NoError(t, err, "log_level %q should be valid", level)
	}
}

func TestValidateValidLogOutputs(t *testing.T) {
	for _, output := range []string{"console", "file", "both"} {
		cfg := &Config{
			LogOutput: output,
			Connections: map[string]ConnectionConfig{
				"default": {
					Driver:   "postgres",
					Host:     "localhost",
					Database: "testdb",
				},
			},
		}
		err := cfg.Validate()
		assert.NoError(t, err, "log_output %q should be valid", output)
	}
}

func TestValidateValidDrivers(t *testing.T) {
	for _, driver := range []string{"postgres", "mysql", "sqlite"} {
		cfg := &Config{
			Connections: map[string]ConnectionConfig{
				"default": {
					Driver:   driver,
					Host:     "localhost",
					Database: "testdb",
				},
			},
		}
		err := cfg.Validate()
		assert.NoError(t, err, "driver %q should be valid", driver)
	}
}
