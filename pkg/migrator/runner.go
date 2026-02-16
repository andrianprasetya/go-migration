package migrator

import (
	"database/sql"
	"fmt"

	"github.com/andrianprasetya/go-migration/pkg/schema"
)

// Logger defines a minimal logging interface for the runner.
// This will be replaced by the full logger package once implemented.
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// Runner executes migration Up/Down methods within transactions.
type Runner struct {
	db      *sql.DB
	grammar schema.Grammar
	logger  Logger
}

// NewRunner creates a new Runner with the given database connection, grammar, and logger.
// The logger parameter may be nil, in which case logging is silently skipped.
func NewRunner(db *sql.DB, grammar schema.Grammar, logger Logger) *Runner {
	return &Runner{
		db:      db,
		grammar: grammar,
		logger:  logger,
	}
}

// Execute runs a migration in the given direction ("up" or "down").
// If the migration implements TransactionOption and DisableTransaction() returns true,
// it executes without a transaction. Otherwise it delegates to ExecuteInTransaction.
func (r *Runner) Execute(m Migration, direction string) error {
	if opt, ok := m.(TransactionOption); ok && opt.DisableTransaction() {
		return r.executeWithoutTransaction(m, direction)
	}
	return r.ExecuteInTransaction(m, direction)
}

// ExecuteInTransaction runs a migration within a database transaction.
// On success the transaction is committed. On migration error the transaction
// is rolled back. If commit fails, a rollback is attempted and ErrTransactionFailed
// is returned.
func (r *Runner) ExecuteInTransaction(m Migration, direction string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	builder := schema.NewBuilder(tx, r.grammar)

	if err := r.runMigration(m, builder, direction); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback after migration error: %v, original: %w", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback after commit failure: %v: %w", rbErr, ErrTransactionFailed)
		}
		return fmt.Errorf("commit: %w", ErrTransactionFailed)
	}

	return nil
}

// executeWithoutTransaction runs a migration directly against the database
// connection without wrapping it in a transaction.
func (r *Runner) executeWithoutTransaction(m Migration, direction string) error {
	builder := schema.NewBuilder(r.db, r.grammar)
	return r.runMigration(m, builder, direction)
}

// runMigration calls the appropriate migration method based on direction.
func (r *Runner) runMigration(m Migration, builder *schema.Builder, direction string) error {
	switch direction {
	case "up":
		return m.Up(builder)
	case "down":
		return m.Down(builder)
	default:
		return fmt.Errorf("unknown migration direction: %q", direction)
	}
}
