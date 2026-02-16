package migrator

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
)

// Feature: go-migration, Property 20: Hooks are invoked with correct arguments
// **Validates: Requirements 8.2, 8.3**

// genMigrationName generates a valid migration name matching ^\d{14}_[a-z][a-z0-9_]*$
func genMigrationName() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		timestamp := rapid.StringMatching(`^[0-9]{14}$`).Draw(t, "timestamp")
		suffix := rapid.StringMatching(`^[a-z][a-z0-9_]{0,20}$`).Draw(t, "suffix")
		return fmt.Sprintf("%s_%s", timestamp, suffix)
	})
}

// genDirection generates "up" or "down".
func genDirection() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{"up", "down"})
}

// TestProperty20_BeforeHooksReceiveCorrectArguments verifies that all registered
// before hooks are called with the exact migration name and direction.
func TestProperty20_BeforeHooksReceiveCorrectArguments(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		name := genMigrationName().Draw(t, "migrationName")
		direction := genDirection().Draw(t, "direction")
		hookCount := rapid.IntRange(1, 10).Draw(t, "hookCount")

		hm := NewHookManager()

		type call struct {
			name      string
			direction string
		}
		var calls []call

		for i := 0; i < hookCount; i++ {
			hm.RegisterBefore(func(n string, d string) error {
				calls = append(calls, call{name: n, direction: d})
				return nil
			})
		}

		err := hm.RunBefore(name, direction)
		assert.NoError(t, err)

		// All before hooks should have been called
		assert.Len(t, calls, hookCount, "all before hooks should be invoked")

		// Each hook should receive the exact migration name and direction
		for i, c := range calls {
			assert.Equal(t, name, c.name, "before hook %d should receive migration name", i)
			assert.Equal(t, direction, c.direction, "before hook %d should receive direction", i)
		}
	})
}

// TestProperty20_AfterHooksReceiveCorrectArguments verifies that all registered
// after hooks are called with the exact migration name, direction, and a non-negative duration.
func TestProperty20_AfterHooksReceiveCorrectArguments(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		name := genMigrationName().Draw(t, "migrationName")
		direction := genDirection().Draw(t, "direction")
		hookCount := rapid.IntRange(1, 10).Draw(t, "hookCount")
		durationMs := rapid.Int64Range(0, 60000).Draw(t, "durationMs")
		dur := time.Duration(durationMs) * time.Millisecond

		hm := NewHookManager()

		type call struct {
			name      string
			direction string
			duration  time.Duration
		}
		var calls []call

		for i := 0; i < hookCount; i++ {
			hm.RegisterAfter(func(n string, d string, dur time.Duration) error {
				calls = append(calls, call{name: n, direction: d, duration: dur})
				return nil
			})
		}

		err := hm.RunAfter(name, direction, dur)
		assert.NoError(t, err)

		// All after hooks should have been called
		assert.Len(t, calls, hookCount, "all after hooks should be invoked")

		// Each hook should receive the exact arguments
		for i, c := range calls {
			assert.Equal(t, name, c.name, "after hook %d should receive migration name", i)
			assert.Equal(t, direction, c.direction, "after hook %d should receive direction", i)
			assert.Equal(t, dur, c.duration, "after hook %d should receive the provided duration", i)
			assert.GreaterOrEqual(t, c.duration, time.Duration(0), "after hook %d duration should be non-negative", i)
		}
	})
}

// TestProperty20_HooksCalledInRegistrationOrder verifies that both before and after
// hooks are invoked in the order they were registered.
func TestProperty20_HooksCalledInRegistrationOrder(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		name := genMigrationName().Draw(t, "migrationName")
		direction := genDirection().Draw(t, "direction")
		beforeCount := rapid.IntRange(1, 10).Draw(t, "beforeCount")
		afterCount := rapid.IntRange(1, 10).Draw(t, "afterCount")
		durationMs := rapid.Int64Range(0, 60000).Draw(t, "durationMs")
		dur := time.Duration(durationMs) * time.Millisecond

		hm := NewHookManager()

		var beforeOrder []int
		var afterOrder []int

		for i := 0; i < beforeCount; i++ {
			idx := i
			hm.RegisterBefore(func(n string, d string) error {
				beforeOrder = append(beforeOrder, idx)
				return nil
			})
		}

		for i := 0; i < afterCount; i++ {
			idx := i
			hm.RegisterAfter(func(n string, d string, dur time.Duration) error {
				afterOrder = append(afterOrder, idx)
				return nil
			})
		}

		err := hm.RunBefore(name, direction)
		assert.NoError(t, err)

		err = hm.RunAfter(name, direction, dur)
		assert.NoError(t, err)

		// Verify before hooks ran in registration order
		assert.Len(t, beforeOrder, beforeCount)
		for i := 0; i < beforeCount; i++ {
			assert.Equal(t, i, beforeOrder[i], "before hook %d should run in registration order", i)
		}

		// Verify after hooks ran in registration order
		assert.Len(t, afterOrder, afterCount)
		for i := 0; i < afterCount; i++ {
			assert.Equal(t, i, afterOrder[i], "after hook %d should run in registration order", i)
		}
	})
}
