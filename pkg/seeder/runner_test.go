package seeder

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// trackingSeeder records when it was run for order verification.
type trackingSeeder struct {
	name  string
	order *[]string
}

func (s *trackingSeeder) Run(db *sql.DB) error {
	*s.order = append(*s.order, s.name)
	return nil
}

// dependentTrackingSeeder is a seeder with dependencies that tracks execution.
type dependentTrackingSeeder struct {
	name  string
	deps  []string
	order *[]string
}

func (s *dependentTrackingSeeder) Run(db *sql.DB) error {
	*s.order = append(*s.order, s.name)
	return nil
}

func (s *dependentTrackingSeeder) DependsOn() []string {
	return s.deps
}

// failingSeeder always returns an error.
type failingSeeder struct {
	name string
	err  error
}

func (s *failingSeeder) Run(db *sql.DB) error {
	return s.err
}

// testLogger captures log messages for verification.
type testLogger struct {
	infos  []string
	errors []string
}

func (l *testLogger) Info(msg string, args ...interface{}) {
	l.infos = append(l.infos, fmt.Sprintf(msg, args...))
}

func (l *testLogger) Error(msg string, args ...interface{}) {
	l.errors = append(l.errors, fmt.Sprintf(msg, args...))
}

func newTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func TestNewRunner(t *testing.T) {
	reg := NewRegistry()
	db, _ := newTestDB(t)
	logger := &testLogger{}

	runner := NewRunner(reg, db, logger)
	assert.NotNil(t, runner)
}

func TestRunAllNoDependencies(t *testing.T) {
	reg := NewRegistry()
	db, _ := newTestDB(t)
	var order []string

	require.NoError(t, reg.Register("users", &trackingSeeder{name: "users", order: &order}))
	require.NoError(t, reg.Register("posts", &trackingSeeder{name: "posts", order: &order}))
	require.NoError(t, reg.Register("comments", &trackingSeeder{name: "comments", order: &order}))

	runner := NewRunner(reg, db, nil)
	err := runner.RunAll()
	require.NoError(t, err)

	// All seeders should have been executed
	assert.Len(t, order, 3)
	assert.Contains(t, order, "users")
	assert.Contains(t, order, "posts")
	assert.Contains(t, order, "comments")
}

func TestRunAllWithDependencies(t *testing.T) {
	reg := NewRegistry()
	db, _ := newTestDB(t)
	var order []string

	// comments depends on posts, posts depends on users
	require.NoError(t, reg.Register("users", &trackingSeeder{name: "users", order: &order}))
	require.NoError(t, reg.Register("posts", &dependentTrackingSeeder{
		name: "posts", deps: []string{"users"}, order: &order,
	}))
	require.NoError(t, reg.Register("comments", &dependentTrackingSeeder{
		name: "comments", deps: []string{"posts"}, order: &order,
	}))

	runner := NewRunner(reg, db, nil)
	err := runner.RunAll()
	require.NoError(t, err)

	assert.Len(t, order, 3)
	// users must come before posts, posts must come before comments
	usersIdx := indexOf(order, "users")
	postsIdx := indexOf(order, "posts")
	commentsIdx := indexOf(order, "comments")
	assert.Less(t, usersIdx, postsIdx, "users should run before posts")
	assert.Less(t, postsIdx, commentsIdx, "posts should run before comments")
}

func TestRunSingleWithDependencies(t *testing.T) {
	reg := NewRegistry()
	db, _ := newTestDB(t)
	var order []string

	require.NoError(t, reg.Register("users", &trackingSeeder{name: "users", order: &order}))
	require.NoError(t, reg.Register("posts", &dependentTrackingSeeder{
		name: "posts", deps: []string{"users"}, order: &order,
	}))
	require.NoError(t, reg.Register("comments", &dependentTrackingSeeder{
		name: "comments", deps: []string{"posts"}, order: &order,
	}))

	runner := NewRunner(reg, db, nil)
	// Running "comments" should also run users and posts first
	err := runner.Run("comments")
	require.NoError(t, err)

	assert.Equal(t, []string{"users", "posts", "comments"}, order)
}

func TestRunSingleNoDependencies(t *testing.T) {
	reg := NewRegistry()
	db, _ := newTestDB(t)
	var order []string

	require.NoError(t, reg.Register("users", &trackingSeeder{name: "users", order: &order}))
	require.NoError(t, reg.Register("posts", &trackingSeeder{name: "posts", order: &order}))

	runner := NewRunner(reg, db, nil)
	err := runner.Run("users")
	require.NoError(t, err)

	// Only "users" should have run
	assert.Equal(t, []string{"users"}, order)
}

func TestCircularDependencyDetection(t *testing.T) {
	reg := NewRegistry()
	db, _ := newTestDB(t)
	var order []string

	// A -> B -> C -> A (cycle)
	require.NoError(t, reg.Register("a", &dependentTrackingSeeder{
		name: "a", deps: []string{"c"}, order: &order,
	}))
	require.NoError(t, reg.Register("b", &dependentTrackingSeeder{
		name: "b", deps: []string{"a"}, order: &order,
	}))
	require.NoError(t, reg.Register("c", &dependentTrackingSeeder{
		name: "c", deps: []string{"b"}, order: &order,
	}))

	runner := NewRunner(reg, db, nil)

	// RunAll should detect the cycle
	err := runner.RunAll()
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrCircularDependency))

	// Run single should also detect the cycle
	order = nil
	err = runner.Run("a")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrCircularDependency))
}

func TestSeederFailureStopsExecution(t *testing.T) {
	reg := NewRegistry()
	db, _ := newTestDB(t)
	var order []string
	seederErr := fmt.Errorf("seed failed")

	require.NoError(t, reg.Register("users", &trackingSeeder{name: "users", order: &order}))
	require.NoError(t, reg.Register("posts", &failingSeeder{name: "posts", err: seederErr}))
	require.NoError(t, reg.Register("tags", &trackingSeeder{name: "tags", order: &order}))

	// posts depends on users, tags depends on posts
	// Replace with dependent seeders to ensure order
	reg2 := NewRegistry()
	require.NoError(t, reg2.Register("users", &trackingSeeder{name: "users", order: &order}))
	require.NoError(t, reg2.Register("posts", &dependentTrackingSeeder{
		name: "posts", deps: []string{"users"}, order: &order,
	}))

	// Use a simpler scenario: users succeeds, then failing_seeder fails
	reg3 := NewRegistry()
	var order3 []string
	require.NoError(t, reg3.Register("aaa_first", &trackingSeeder{name: "aaa_first", order: &order3}))
	require.NoError(t, reg3.Register("bbb_failing", &failingSeeder{name: "bbb_failing", err: seederErr}))
	require.NoError(t, reg3.Register("ccc_last", &trackingSeeder{name: "ccc_last", order: &order3}))

	runner := NewRunner(reg3, db, nil)
	err := runner.RunAll()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bbb_failing")
	assert.ErrorIs(t, err, seederErr)

	// "aaa_first" should have run, "ccc_last" should NOT have run
	assert.Contains(t, order3, "aaa_first")
	assert.NotContains(t, order3, "ccc_last")
}

func TestRunNonExistentSeeder(t *testing.T) {
	reg := NewRegistry()
	db, _ := newTestDB(t)

	runner := NewRunner(reg, db, nil)
	err := runner.Run("nonexistent")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSeederNotFound))
}

func TestRunAllEmptyRegistry(t *testing.T) {
	reg := NewRegistry()
	db, _ := newTestDB(t)

	runner := NewRunner(reg, db, nil)
	err := runner.RunAll()
	require.NoError(t, err)
}

func TestRunWithLogger(t *testing.T) {
	reg := NewRegistry()
	db, _ := newTestDB(t)
	logger := &testLogger{}
	var order []string

	require.NoError(t, reg.Register("users", &trackingSeeder{name: "users", order: &order}))

	runner := NewRunner(reg, db, logger)
	err := runner.RunAll()
	require.NoError(t, err)

	assert.NotEmpty(t, logger.infos)
}

func TestRunWithNilLogger(t *testing.T) {
	reg := NewRegistry()
	db, _ := newTestDB(t)
	var order []string

	require.NoError(t, reg.Register("users", &trackingSeeder{name: "users", order: &order}))

	runner := NewRunner(reg, db, nil)
	// Should not panic with nil logger
	err := runner.RunAll()
	require.NoError(t, err)
}

func TestCircularDependencyErrorMessage(t *testing.T) {
	reg := NewRegistry()
	db, _ := newTestDB(t)
	var order []string

	// Self-referencing seeder
	require.NoError(t, reg.Register("self", &dependentTrackingSeeder{
		name: "self", deps: []string{"self"}, order: &order,
	}))

	runner := NewRunner(reg, db, nil)
	err := runner.Run("self")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrCircularDependency))
	assert.Contains(t, err.Error(), "self")
}

func TestDiamondDependency(t *testing.T) {
	reg := NewRegistry()
	db, _ := newTestDB(t)
	var order []string

	// Diamond: D depends on B and C, both B and C depend on A
	require.NoError(t, reg.Register("a", &trackingSeeder{name: "a", order: &order}))
	require.NoError(t, reg.Register("b", &dependentTrackingSeeder{
		name: "b", deps: []string{"a"}, order: &order,
	}))
	require.NoError(t, reg.Register("c", &dependentTrackingSeeder{
		name: "c", deps: []string{"a"}, order: &order,
	}))
	require.NoError(t, reg.Register("d", &dependentTrackingSeeder{
		name: "d", deps: []string{"b", "c"}, order: &order,
	}))

	runner := NewRunner(reg, db, nil)
	err := runner.Run("d")
	require.NoError(t, err)

	// A should run first, then B and C (in some order), then D last
	assert.Len(t, order, 4)
	assert.Equal(t, "a", order[0], "a should run first")
	assert.Equal(t, "d", order[3], "d should run last")
	// Each seeder should run exactly once
	assert.ElementsMatch(t, []string{"a", "b", "c", "d"}, order)
}

func TestSeederFailureInRunStopsExecution(t *testing.T) {
	reg := NewRegistry()
	db, _ := newTestDB(t)
	var order []string
	seederErr := fmt.Errorf("users seed failed")

	// posts depends on users, users fails
	require.NoError(t, reg.Register("users", &failingSeeder{name: "users", err: seederErr}))
	require.NoError(t, reg.Register("posts", &dependentTrackingSeeder{
		name: "posts", deps: []string{"users"}, order: &order,
	}))

	runner := NewRunner(reg, db, nil)
	err := runner.Run("posts")
	require.Error(t, err)
	assert.ErrorIs(t, err, seederErr)
	assert.Contains(t, err.Error(), "users")
	// posts should not have run
	assert.Empty(t, order)
}

// indexOf returns the index of s in slice, or -1 if not found.
func indexOf(slice []string, s string) int {
	for i, v := range slice {
		if v == s {
			return i
		}
	}
	return -1
}
