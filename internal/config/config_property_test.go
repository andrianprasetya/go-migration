package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"pgregory.net/rapid"
)

// connConfigGenerator returns a rapid generator for ConnectionConfig with all required fields populated.
func connConfigGenerator() *rapid.Generator[ConnectionConfig] {
	return rapid.Custom(func(t *rapid.T) ConnectionConfig {
		drivers := []string{"postgres", "mysql", "sqlite"}
		return ConnectionConfig{
			Driver:   rapid.SampledFrom(drivers).Draw(t, "driver"),
			Host:     rapid.StringMatching(`[a-z]{3,10}\.[a-z]{2,5}`).Draw(t, "host"),
			Port:     rapid.IntRange(1, 65535).Draw(t, "port"),
			Database: rapid.StringMatching(`[a-z]{3,12}`).Draw(t, "database"),
			Username: rapid.StringMatching(`[a-z]{3,10}`).Draw(t, "username"),
			Password: rapid.StringMatching(`[a-zA-Z0-9]{5,15}`).Draw(t, "password"),
		}
	})
}

// validConfigGenerator returns a rapid generator for Config with at least one valid connection.
func validConfigGenerator() *rapid.Generator[Config] {
	return rapid.Custom(func(t *rapid.T) Config {
		numConns := rapid.IntRange(1, 3).Draw(t, "numConns")
		conns := make(map[string]ConnectionConfig, numConns)
		var firstName string
		for i := 0; i < numConns; i++ {
			name := rapid.StringMatching(`[a-z]{3,8}`).Draw(t, "connName")
			if i == 0 {
				firstName = name
			}
			conns[name] = connConfigGenerator().Draw(t, "conn")
		}
		return Config{
			Connections:    conns,
			DefaultConn:    firstName,
			MigrationTable: rapid.StringMatching(`[a-z_]{5,15}`).Draw(t, "migrationTable"),
			MigrationDir:   rapid.StringMatching(`[a-z/]{5,15}`).Draw(t, "migrationDir"),
			SeederDir:      rapid.StringMatching(`[a-z/]{5,15}`).Draw(t, "seederDir"),
			LogLevel:       rapid.SampledFrom([]string{"debug", "info", "warn", "error"}).Draw(t, "logLevel"),
			LogOutput:      rapid.SampledFrom([]string{"console", "file"}).Draw(t, "logOutput"),
		}
	})
}

// Feature: go-migration, Property 35: Config round-trip across formats
// **Validates: Requirements 15.1, 15.2**
//
// For any valid Config struct, serializing to YAML and loading back should
// produce an equivalent Config. Similarly for JSON.
func TestProperty35_ConfigRoundTripAcrossFormats(t *testing.T) {
	t.Run("YAML round-trip", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			original := validConfigGenerator().Draw(t, "config")

			// Serialize to YAML
			data, err := yaml.Marshal(&original)
			require.NoError(t, err, "YAML marshal should succeed")

			// Write to temp file
			dir, err := os.MkdirTemp("", "config-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(dir)

			path := filepath.Join(dir, "config.yaml")
			err = os.WriteFile(path, data, 0o644)
			require.NoError(t, err)

			// Load back
			loaded, err := Load(path)
			require.NoError(t, err, "Load should succeed for valid YAML")

			// Compare connections (Load applies defaults, so compare connections directly)
			assert.Equal(t, len(original.Connections), len(loaded.Connections),
				"Should have same number of connections")

			for name, origConn := range original.Connections {
				loadedConn, ok := loaded.Connections[name]
				assert.True(t, ok, "Connection %q should exist after round-trip", name)
				assert.Equal(t, origConn.Driver, loadedConn.Driver)
				assert.Equal(t, origConn.Host, loadedConn.Host)
				assert.Equal(t, origConn.Port, loadedConn.Port)
				assert.Equal(t, origConn.Database, loadedConn.Database)
				assert.Equal(t, origConn.Username, loadedConn.Username)
				assert.Equal(t, origConn.Password, loadedConn.Password)
			}

			// Compare top-level fields
			assert.Equal(t, original.DefaultConn, loaded.DefaultConn)
			assert.Equal(t, original.MigrationTable, loaded.MigrationTable)
			assert.Equal(t, original.MigrationDir, loaded.MigrationDir)
			assert.Equal(t, original.SeederDir, loaded.SeederDir)
			assert.Equal(t, original.LogLevel, loaded.LogLevel)
			assert.Equal(t, original.LogOutput, loaded.LogOutput)
		})
	})

	t.Run("JSON round-trip", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			original := validConfigGenerator().Draw(t, "config")

			// Serialize to JSON
			data, err := json.Marshal(&original)
			require.NoError(t, err, "JSON marshal should succeed")

			// Write to temp file
			dir, err := os.MkdirTemp("", "config-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(dir)

			path := filepath.Join(dir, "config.json")
			err = os.WriteFile(path, data, 0o644)
			require.NoError(t, err)

			// Load back
			loaded, err := Load(path)
			require.NoError(t, err, "Load should succeed for valid JSON")

			// Compare connections
			assert.Equal(t, len(original.Connections), len(loaded.Connections),
				"Should have same number of connections")

			for name, origConn := range original.Connections {
				loadedConn, ok := loaded.Connections[name]
				assert.True(t, ok, "Connection %q should exist after round-trip", name)
				assert.Equal(t, origConn.Driver, loadedConn.Driver)
				assert.Equal(t, origConn.Host, loadedConn.Host)
				assert.Equal(t, origConn.Port, loadedConn.Port)
				assert.Equal(t, origConn.Database, loadedConn.Database)
				assert.Equal(t, origConn.Username, loadedConn.Username)
				assert.Equal(t, origConn.Password, loadedConn.Password)
			}

			// Compare top-level fields
			assert.Equal(t, original.DefaultConn, loaded.DefaultConn)
			assert.Equal(t, original.MigrationTable, loaded.MigrationTable)
			assert.Equal(t, original.MigrationDir, loaded.MigrationDir)
			assert.Equal(t, original.SeederDir, loaded.SeederDir)
			assert.Equal(t, original.LogLevel, loaded.LogLevel)
			assert.Equal(t, original.LogOutput, loaded.LogOutput)
		})
	})
}

// Feature: go-migration, Property 36: Missing required fields produce validation errors
// **Validates: Requirements 15.4**
//
// For any Config missing one or more required fields (driver, host, database
// for each connection), Validate() should return an error listing all missing fields.
func TestProperty36_MissingRequiredFieldsProduceValidationErrors(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Start with a valid config
		cfg := validConfigGenerator().Draw(t, "config")

		// Pick a connection to corrupt
		var connNames []string
		for name := range cfg.Connections {
			connNames = append(connNames, name)
		}
		targetName := rapid.SampledFrom(connNames).Draw(t, "targetConn")
		conn := cfg.Connections[targetName]

		// Randomly remove one or more required fields
		removeDriver := rapid.Bool().Draw(t, "removeDriver")
		removeHost := rapid.Bool().Draw(t, "removeHost")
		removeDatabase := rapid.Bool().Draw(t, "removeDatabase")

		// Ensure at least one field is removed
		if !removeDriver && !removeHost && !removeDatabase {
			removeDriver = true
		}

		var expectedMissing []string
		if removeDriver {
			conn.Driver = ""
			expectedMissing = append(expectedMissing, "connections."+targetName+".driver")
		}
		if removeHost {
			conn.Host = ""
			expectedMissing = append(expectedMissing, "connections."+targetName+".host")
		}
		if removeDatabase {
			conn.Database = ""
			expectedMissing = append(expectedMissing, "connections."+targetName+".database")
		}
		cfg.Connections[targetName] = conn

		// Validate should return an error
		err := cfg.Validate()
		require.Error(t, err, "Validate should fail when required fields are missing")
		assert.ErrorIs(t, err, ErrConfigValidation, "Error should wrap ErrConfigValidation")

		// Error message should mention all missing fields
		errMsg := err.Error()
		for _, field := range expectedMissing {
			assert.True(t, strings.Contains(errMsg, field),
				"Error should mention missing field %q, got: %s", field, errMsg)
		}
	})
}
