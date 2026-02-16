package main

import (
	"fmt"
	"os"

	"github.com/andrianprasetya/go-migration/internal/config"
	"github.com/andrianprasetya/go-migration/internal/generator"
	"github.com/andrianprasetya/go-migration/internal/logger"
	"github.com/andrianprasetya/go-migration/pkg/cli"
	"github.com/andrianprasetya/go-migration/pkg/cli/commands"
	"github.com/andrianprasetya/go-migration/pkg/database"
	"github.com/andrianprasetya/go-migration/pkg/database/drivers"
	"github.com/andrianprasetya/go-migration/pkg/migrator"
	"github.com/andrianprasetya/go-migration/pkg/seeder"
	"github.com/spf13/cobra"
)

// commandsNeedingDB lists commands that require a database connection.
// Commands not in this set (version, help, make:migration, make:seeder)
// skip the PersistentPreRunE setup entirely.
var commandsNeedingDB = map[string]bool{
	"migrate":          true,
	"migrate:rollback": true,
	"migrate:reset":    true,
	"migrate:refresh":  true,
	"migrate:fresh":    true,
	"migrate:status":   true,
	"migrate:install":  true,
	"db:seed":          true,
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	app := cli.NewApp(nil)
	root := app.Root()

	// cmdCtx is populated lazily by PersistentPreRunE for commands that need DB.
	var cmdCtx *commands.CommandContext
	// connManager is stored so we can defer Close().
	var connManager *database.Manager

	getCtx := func() *commands.CommandContext {
		return cmdCtx
	}

	// Register all sub-commands.
	root.AddCommand(
		newVersionCommand(),
		commands.NewMigrateCommand(getCtx),
		commands.NewMigrateRollbackCommand(getCtx),
		commands.NewMigrateResetCommand(getCtx),
		commands.NewMigrateRefreshCommand(getCtx),
		commands.NewMigrateFreshCommand(getCtx),
		commands.NewMigrateStatusCommand(getCtx),
		commands.NewMigrateInstallCommand(getCtx),
		commands.NewMakeMigrationCommand(getCtx),
		commands.NewMakeSeederCommand(getCtx),
		commands.NewSeedCommand(getCtx),
	)

	// PersistentPreRunE initialises config, DB, and services for commands
	// that need a database connection. Lightweight commands (version, help,
	// make:migration, make:seeder) skip this entirely.
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if !commandsNeedingDB[cmd.Name()] {
			// For make:migration / make:seeder we still need the generator,
			// so build a minimal context with just the generator.
			if cmd.Name() == "make:migration" || cmd.Name() == "make:seeder" {
				cfg, err := loadConfig(cmd)
				if err != nil {
					return err
				}
				gen := generator.NewGenerator(cfg.MigrationDir)
				if cmd.Name() == "make:seeder" {
					gen = generator.NewGenerator(cfg.SeederDir)
				}
				cmdCtx = &commands.CommandContext{Generator: gen}
			}
			return nil
		}

		// Load configuration.
		cfg, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		// Set up logger.
		log := setupLogger(cfg)

		// Set up database connection manager with all drivers.
		connManager = database.NewManager()
		connManager.RegisterDriver("postgres", drivers.NewPostgresDriver())
		connManager.RegisterDriver("mysql", drivers.NewMySQLDriver())
		connManager.RegisterDriver("sqlite3", drivers.NewSQLiteDriver())

		// Add all configured connections.
		for name, connCfg := range cfg.Connections {
			dbCfg := toDBConnectionConfig(connCfg)
			if err := connManager.AddConnection(name, dbCfg); err != nil {
				return fmt.Errorf("add connection %q: %w", name, err)
			}
		}

		// Set default connection if specified.
		if cfg.DefaultConn != "" {
			if err := connManager.SetDefault(cfg.DefaultConn); err != nil {
				return fmt.Errorf("set default connection: %w", err)
			}
		}

		// Get the default DB connection.
		db, err := connManager.Default()
		if err != nil {
			return fmt.Errorf("get default connection: %w", err)
		}

		// Create Migrator.
		m := migrator.New(db,
			migrator.WithTableName(cfg.MigrationTable),
			migrator.WithLogger(log),
		)

		// Create Seeder Runner.
		seederRegistry := seeder.NewRegistry()
		seederRunner := seeder.NewRunner(seederRegistry, db, log)

		// Create Generator.
		gen := generator.NewGenerator(cfg.MigrationDir)

		cmdCtx = &commands.CommandContext{
			DB:        db,
			Migrator:  m,
			Seeder:    seederRunner,
			Generator: gen,
		}

		return nil
	}

	// Run the CLI.
	err := app.Run(nil)

	// Cleanup: close all database connections.
	if connManager != nil {
		connManager.Close()
	}

	return err
}

// loadConfig reads the --config flag and loads configuration.
// It tries the file first, then falls back to environment variables.
func loadConfig(cmd *cobra.Command) (*config.Config, error) {
	configPath, _ := cmd.Flags().GetString("config")

	cfg, err := config.Load(configPath)
	if err != nil {
		// Fall back to environment variables.
		cfg, err = config.LoadFromEnv()
		if err != nil {
			return nil, fmt.Errorf("load config: %w", err)
		}
	}

	cfg.ApplyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// setupLogger creates a logger based on the configuration.
func setupLogger(cfg *config.Config) logger.Logger {
	level := logger.ParseLevel(cfg.LogLevel)

	if cfg.LogOutput != "" && cfg.LogOutput != "console" {
		fl, err := logger.NewFileLogger(cfg.LogOutput, level)
		if err == nil {
			return fl
		}
		// Fall back to console if file logger fails.
	}

	return logger.NewConsoleLogger(level)
}

// toDBConnectionConfig converts a config.ConnectionConfig to a database.ConnectionConfig.
func toDBConnectionConfig(c config.ConnectionConfig) database.ConnectionConfig {
	return database.ConnectionConfig{
		Driver:          c.Driver,
		Host:            c.Host,
		Port:            c.Port,
		Database:        c.Database,
		Username:        c.Username,
		Password:        c.Password,
		MaxOpenConns:    c.MaxOpenConns,
		MaxIdleConns:    c.MaxIdleConns,
		ConnMaxLifetime: c.ConnMaxLifetime,
		Options:         c.Options,
	}
}
