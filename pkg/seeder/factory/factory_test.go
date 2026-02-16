package factory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// User is a simple test struct for factory tests.
type User struct {
	Name  string
	Email string
	Age   int
}

func userFactory() *Factory[User] {
	faker := NewFaker(42)
	return NewFactory[User](func(f Faker) User {
		return User{
			Name:  f.Name(),
			Email: f.Email(),
			Age:   f.IntBetween(18, 65),
		}
	}).WithFaker(faker)
}

func TestMake_ReturnsPopulatedInstance(t *testing.T) {
	f := userFactory()
	user := f.Make()

	assert.NotEmpty(t, user.Name, "Name should be populated")
	assert.NotEmpty(t, user.Email, "Email should be populated")
	assert.True(t, user.Age >= 18 && user.Age <= 65, "Age should be in range [18,65]")
}

func TestMakeMany_ReturnsCorrectCount(t *testing.T) {
	f := userFactory()

	users := f.MakeMany(5)
	require.Len(t, users, 5)

	for _, u := range users {
		assert.NotEmpty(t, u.Name)
		assert.NotEmpty(t, u.Email)
	}
}

func TestMakeMany_ZeroReturnsNil(t *testing.T) {
	f := userFactory()
	assert.Nil(t, f.MakeMany(0))
}

func TestMakeMany_NegativeReturnsNil(t *testing.T) {
	f := userFactory()
	assert.Nil(t, f.MakeMany(-1))
}

func TestState_OverridesSpecificFields(t *testing.T) {
	f := userFactory()
	f.State("admin", func(faker Faker, base User) User {
		base.Name = "Admin User"
		return base
	})

	admin := f.WithState("admin").Make()
	assert.Equal(t, "Admin User", admin.Name, "State should override Name")
	assert.NotEmpty(t, admin.Email, "Email should still be populated from definition")
	assert.True(t, admin.Age >= 18 && admin.Age <= 65, "Age should still be in range")
}

func TestWithState_DoesNotModifyOriginalFactory(t *testing.T) {
	f := userFactory()
	f.State("senior", func(faker Faker, base User) User {
		base.Age = 99
		return base
	})

	withSenior := f.WithState("senior")

	// Original factory should produce normal instances
	original := f.Make()
	assert.True(t, original.Age >= 18 && original.Age <= 65,
		"Original factory should not be affected by WithState")

	// Derived factory should apply the state
	senior := withSenior.Make()
	assert.Equal(t, 99, senior.Age, "WithState factory should apply state override")
}

func TestWithState_UnknownStateIsIgnored(t *testing.T) {
	f := userFactory()
	// Applying a state that was never registered should not panic
	user := f.WithState("nonexistent").Make()
	assert.NotEmpty(t, user.Name)
}

func TestMultipleStates_AppliedInOrder(t *testing.T) {
	f := userFactory()
	f.State("named", func(faker Faker, base User) User {
		base.Name = "Custom Name"
		return base
	})
	f.State("aged", func(faker Faker, base User) User {
		base.Age = 100
		return base
	})

	user := f.WithState("named").WithState("aged").Make()
	assert.Equal(t, "Custom Name", user.Name)
	assert.Equal(t, 100, user.Age)
}

func TestWithFaker_UsesProvidedFaker(t *testing.T) {
	f := NewFactory[User](func(faker Faker) User {
		return User{
			Name:  faker.Name(),
			Email: faker.Email(),
			Age:   faker.IntBetween(18, 65),
		}
	})

	// Two factories with the same seed should produce identical results
	f1 := f.WithFaker(NewFaker(123))
	f2 := f.WithFaker(NewFaker(123))

	u1 := f1.Make()
	u2 := f2.Make()

	assert.Equal(t, u1.Name, u2.Name)
	assert.Equal(t, u1.Email, u2.Email)
	assert.Equal(t, u1.Age, u2.Age)
}

func TestStateChaining(t *testing.T) {
	f := NewFactory[User](func(faker Faker) User {
		return User{Name: faker.Name(), Email: faker.Email(), Age: faker.IntBetween(18, 65)}
	}).WithFaker(NewFaker(42))

	// State() returns the same factory for chaining
	result := f.State("a", func(faker Faker, base User) User {
		return base
	}).State("b", func(faker Faker, base User) User {
		return base
	})

	assert.Equal(t, f, result, "State() should return the same factory for chaining")
}
