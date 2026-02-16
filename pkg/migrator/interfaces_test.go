package migrator_test

import (
	"testing"

	"github.com/andrianprasetya/go-migration/pkg/migrator"
	"github.com/andrianprasetya/go-migration/pkg/schema"
	"github.com/stretchr/testify/assert"
)

// stubMigration is a minimal Migration implementation for testing.
type stubMigration struct {
	upCalled   bool
	downCalled bool
}

func (s *stubMigration) Up(b *schema.Builder) error {
	s.upCalled = true
	return nil
}

func (s *stubMigration) Down(b *schema.Builder) error {
	s.downCalled = true
	return nil
}

// stubNoTxMigration implements both Migration and TransactionOption.
type stubNoTxMigration struct {
	stubMigration
}

func (s *stubNoTxMigration) DisableTransaction() bool {
	return true
}

func TestMigrationInterfaceCompliance(t *testing.T) {
	var m migrator.Migration = &stubMigration{}
	assert.NotNil(t, m)

	err := m.Up(&schema.Builder{})
	assert.NoError(t, err)

	err = m.Down(&schema.Builder{})
	assert.NoError(t, err)
}

func TestTransactionOptionInterfaceCompliance(t *testing.T) {
	s := &stubNoTxMigration{}

	// Verify it satisfies both interfaces.
	var m migrator.Migration = s
	var opt migrator.TransactionOption = s

	assert.NotNil(t, m)
	assert.NotNil(t, opt)
	assert.True(t, opt.DisableTransaction())
}

func TestMigrationWithoutTransactionOption(t *testing.T) {
	s := &stubMigration{}

	// A plain Migration should NOT satisfy TransactionOption.
	var m migrator.Migration = s
	_, ok := m.(migrator.TransactionOption)
	assert.False(t, ok, "stubMigration should not implement TransactionOption")
}
