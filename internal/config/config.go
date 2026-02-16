// Package config provides configuration loading and validation for go-migration.
// It supports YAML files, JSON files, and environment variables.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ErrConfigValidation is returned when configuration validation fails.
var ErrConfigValidation = errors.New("configuration validation failed")

// Config holds the top-level configuration for go-migration.
type Config struct {
	Connections    map[string]ConnectionConfig `yaml:"connections" json:"connections"`
	DefaultConn    string                      `yaml:"default" json:"default"`
	MigrationTable string                      `yaml:"migration_table" json:"migration_table"`
	MigrationDir   string                      `yaml:"migration_dir" json:"migration_dir"`
	SeederDir      string                      `yaml:"seeder_dir" json:"seeder_dir"`
	LogLevel       string                      `yaml:"log_level" json:"log_level"`
	LogOutput      string                      `yaml:"log_output" json:"log_output"`
}

// ConnectionConfig holds the configuration for a single database connection.
type ConnectionConfig struct {
	Driver          string            `yaml:"driver" json:"driver"`
	Host            string            `yaml:"host" json:"host"`
	Port            int               `yaml:"port" json:"port"`
	Database        string            `yaml:"database" json:"database"`
	Username        string            `yaml:"username" json:"username"`
	Password        string            `yaml:"password" json:"password"`
	MaxOpenConns    int               `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int               `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration     `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
	Options         map[string]string `yaml:"options" json:"options"`
}

// ApplyDefaults sets sensible default values for optional settings.
func (c *Config) ApplyDefaults() {
	if c.MigrationTable == "" {
		c.MigrationTable = "migrations"
	}
	if c.MigrationDir == "" {
		c.MigrationDir = "migrations"
	}
	if c.SeederDir == "" {
		c.SeederDir = "seeders"
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	if c.LogOutput == "" {
		c.LogOutput = "console"
	}
	if c.Connections == nil {
		c.Connections = make(map[string]ConnectionConfig)
	}
}

// Validate checks that the configuration has all required fields.
// Returns an error wrapping ErrConfigValidation listing all missing fields.
func (c *Config) Validate() error {
	var missing []string

	if len(c.Connections) == 0 {
		missing = append(missing, "at least one connection must be defined")
	}

	for name, conn := range c.Connections {
		if conn.Driver == "" {
			missing = append(missing, fmt.Sprintf("connections.%s.driver", name))
		}
		if conn.Host == "" {
			missing = append(missing, fmt.Sprintf("connections.%s.host", name))
		}
		if conn.Database == "" {
			missing = append(missing, fmt.Sprintf("connections.%s.database", name))
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("%w: missing required fields: %s", ErrConfigValidation, strings.Join(missing, ", "))
	}
	return nil
}

// Load reads configuration from a file at the given path.
// It detects the format by file extension (.yaml/.yml for YAML, .json for JSON).
// After loading, it applies defaults.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	cfg := &Config{}
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	cfg.ApplyDefaults()
	return cfg, nil
}

// LoadFromEnv builds a Config from environment variables with the GOMIGRATE_ prefix.
// It creates a single "default" connection from the DB-related env vars.
// After loading, it applies defaults.
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		Connections: make(map[string]ConnectionConfig),
	}

	// Top-level settings
	cfg.DefaultConn = getEnv("GOMIGRATE_DEFAULT_CONNECTION", "default")
	cfg.MigrationTable = getEnv("GOMIGRATE_MIGRATION_TABLE", "")
	cfg.MigrationDir = getEnv("GOMIGRATE_MIGRATION_DIR", "")
	cfg.SeederDir = getEnv("GOMIGRATE_SEEDER_DIR", "")
	cfg.LogLevel = getEnv("GOMIGRATE_LOG_LEVEL", "")
	cfg.LogOutput = getEnv("GOMIGRATE_LOG_OUTPUT", "")

	// Build default connection from env vars
	conn := ConnectionConfig{
		Driver:   getEnv("GOMIGRATE_DB_DRIVER", ""),
		Host:     getEnv("GOMIGRATE_DB_HOST", ""),
		Database: getEnv("GOMIGRATE_DB_DATABASE", ""),
		Username: getEnv("GOMIGRATE_DB_USERNAME", ""),
		Password: getEnv("GOMIGRATE_DB_PASSWORD", ""),
	}

	if portStr := getEnv("GOMIGRATE_DB_PORT", ""); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid GOMIGRATE_DB_PORT value %q: %w", portStr, err)
		}
		conn.Port = port
	}

	// Only add the connection if at least the driver is specified
	if conn.Driver != "" || conn.Host != "" || conn.Database != "" {
		connName := cfg.DefaultConn
		cfg.Connections[connName] = conn
	}

	cfg.ApplyDefaults()
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
