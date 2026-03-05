package migrator

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
)

// Feature: auto-discovery-migrations, Property 1: Sorted order invariant
// For any set of valid migration names registered via AutoRegister in any order,
// GetAutoRegistered() SHALL return them in lexicographic (timestamp) sorted order.
// **Validates: Requirements 1.2, 1.5**
func TestAutoRegistryProperty1_SortedOrderInvariant(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		defer ResetAutoRegistry()

		names := uniqueValidNamesGen().Draw(t, "names")

		// Register in the random order provided by the generator
		for _, name := range names {
			AutoRegister(name, &stubMigration{})
		}

		result := GetAutoRegistered()
		assert.Len(t, result, len(names))

		// Verify strictly ascending order
		for i := 1; i < len(result); i++ {
			assert.True(t, result[i-1].Name < result[i].Name,
				"expected %q < %q at positions %d, %d", result[i-1].Name, result[i].Name, i-1, i)
		}

		// Verify the result matches a standard sort of the input names
		sorted := make([]string, len(names))
		copy(sorted, names)
		sort.Strings(sorted)
		for i, rm := range result {
			assert.Equal(t, sorted[i], rm.Name)
		}
	})
}

// Feature: auto-discovery-migrations, Property 2: Invalid name panic with descriptive message
// For any string that does not match the pattern YYYYMMDDHHMMSS_description,
// calling AutoRegister with that string SHALL panic, and the panic message SHALL contain the invalid name.
// **Validates: Requirements 1.3, 5.1**
func TestAutoRegistryProperty2_InvalidNamePanicWithDescriptiveMessage(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		defer ResetAutoRegistry()

		invalidName := invalidMigrationNameGen().Draw(t, "invalidName")

		assert.PanicsWithValue(t,
			"AutoRegister: migration name \""+invalidName+"\" is invalid (expected YYYYMMDDHHMMSS_description or YYYY_MM_DD_HHMMSS_RRRR_description)",
			func() {
				AutoRegister(invalidName, &stubMigration{})
			},
			"AutoRegister should panic with descriptive message containing the invalid name %q", invalidName,
		)
	})
}

// Feature: auto-discovery-migrations, Property 3: Duplicate name panic
// For any valid migration name, calling AutoRegister twice with the same name
// SHALL panic on the second call, and the panic message SHALL contain the duplicate name.
// **Validates: Requirements 1.4, 5.3**
func TestAutoRegistryProperty3_DuplicateNamePanic(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		defer ResetAutoRegistry()

		name := validMigrationNameGen().Draw(t, "name")

		// First registration should succeed without panic
		AutoRegister(name, &stubMigration{})

		// Second registration with the same name should panic
		assert.PanicsWithValue(t,
			"AutoRegister: duplicate migration name \""+name+"\"",
			func() {
				AutoRegister(name, &stubMigration{})
			},
			"AutoRegister should panic with message containing the duplicate name %q", name,
		)
	})
}
