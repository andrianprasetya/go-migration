package migrator

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHookManager(t *testing.T) {
	hm := NewHookManager()
	require.NotNil(t, hm)
	assert.Empty(t, hm.beforeHooks)
	assert.Empty(t, hm.afterHooks)
}

// --- RegisterBefore and RunBefore ---

func TestRunBefore_SingleHook(t *testing.T) {
	hm := NewHookManager()

	var calledName, calledDir string
	hm.RegisterBefore(func(name, direction string) error {
		calledName = name
		calledDir = direction
		return nil
	})

	err := hm.RunBefore("20240101000000_create_users", "up")
	require.NoError(t, err)
	assert.Equal(t, "20240101000000_create_users", calledName)
	assert.Equal(t, "up", calledDir)
}

func TestRunBefore_ErrorAbortsExecution(t *testing.T) {
	hm := NewHookManager()
	hookErr := errors.New("before hook failed")

	hm.RegisterBefore(func(name, direction string) error {
		return hookErr
	})

	secondCalled := false
	hm.RegisterBefore(func(name, direction string) error {
		secondCalled = true
		return nil
	})

	err := hm.RunBefore("20240101000000_create_users", "up")
	assert.ErrorIs(t, err, hookErr)
	assert.False(t, secondCalled, "second hook should not run after first hook errors")
}

func TestRunBefore_MultipleHooksRunInOrder(t *testing.T) {
	hm := NewHookManager()
	var order []int

	hm.RegisterBefore(func(name, direction string) error {
		order = append(order, 1)
		return nil
	})
	hm.RegisterBefore(func(name, direction string) error {
		order = append(order, 2)
		return nil
	})
	hm.RegisterBefore(func(name, direction string) error {
		order = append(order, 3)
		return nil
	})

	err := hm.RunBefore("20240101000000_test", "down")
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, order)
}

func TestRunBefore_NoHooksReturnsNil(t *testing.T) {
	hm := NewHookManager()
	err := hm.RunBefore("20240101000000_test", "up")
	assert.NoError(t, err)
}

// --- RegisterAfter and RunAfter ---

func TestRunAfter_SingleHook(t *testing.T) {
	hm := NewHookManager()

	var calledName, calledDir string
	var calledDuration time.Duration
	hm.RegisterAfter(func(name, direction string, duration time.Duration) error {
		calledName = name
		calledDir = direction
		calledDuration = duration
		return nil
	})

	dur := 150 * time.Millisecond
	err := hm.RunAfter("20240101000000_create_users", "up", dur)
	require.NoError(t, err)
	assert.Equal(t, "20240101000000_create_users", calledName)
	assert.Equal(t, "up", calledDir)
	assert.Equal(t, dur, calledDuration)
}

func TestRunAfter_ErrorDoesNotPropagate(t *testing.T) {
	hm := NewHookManager()

	hm.RegisterAfter(func(name, direction string, duration time.Duration) error {
		return errors.New("after hook failed")
	})

	err := hm.RunAfter("20240101000000_create_users", "down", time.Second)
	assert.NoError(t, err, "RunAfter should return nil even when hooks error")
}

func TestRunAfter_MultipleHooksRunInOrder(t *testing.T) {
	hm := NewHookManager()
	var order []int

	hm.RegisterAfter(func(name, direction string, duration time.Duration) error {
		order = append(order, 1)
		return nil
	})
	hm.RegisterAfter(func(name, direction string, duration time.Duration) error {
		order = append(order, 2)
		return nil
	})
	hm.RegisterAfter(func(name, direction string, duration time.Duration) error {
		order = append(order, 3)
		return nil
	})

	err := hm.RunAfter("20240101000000_test", "up", time.Second)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, order)
}

func TestRunAfter_ErrorDoesNotStopSubsequentHooks(t *testing.T) {
	hm := NewHookManager()
	var order []int

	hm.RegisterAfter(func(name, direction string, duration time.Duration) error {
		order = append(order, 1)
		return errors.New("hook 1 error")
	})
	hm.RegisterAfter(func(name, direction string, duration time.Duration) error {
		order = append(order, 2)
		return nil
	})
	hm.RegisterAfter(func(name, direction string, duration time.Duration) error {
		order = append(order, 3)
		return errors.New("hook 3 error")
	})

	err := hm.RunAfter("20240101000000_test", "down", time.Second)
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, order, "all after hooks should run even if some error")
}

func TestRunAfter_NoHooksReturnsNil(t *testing.T) {
	hm := NewHookManager()
	err := hm.RunAfter("20240101000000_test", "up", time.Second)
	assert.NoError(t, err)
}

// --- Direction variants ---

func TestHooks_DownDirection(t *testing.T) {
	hm := NewHookManager()

	var beforeDir, afterDir string
	hm.RegisterBefore(func(name, direction string) error {
		beforeDir = direction
		return nil
	})
	hm.RegisterAfter(func(name, direction string, duration time.Duration) error {
		afterDir = direction
		return nil
	})

	err := hm.RunBefore("20240101000000_test", "down")
	require.NoError(t, err)
	assert.Equal(t, "down", beforeDir)

	err = hm.RunAfter("20240101000000_test", "down", time.Second)
	require.NoError(t, err)
	assert.Equal(t, "down", afterDir)
}
