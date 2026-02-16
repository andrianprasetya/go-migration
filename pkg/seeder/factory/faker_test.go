package factory

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newTestFaker() *DefaultFaker {
	return NewFaker(42)
}

func TestNewFaker(t *testing.T) {
	f := NewFaker(42)
	assert.NotNil(t, f)
}

func TestFakerImplementsInterface(t *testing.T) {
	var _ Faker = NewFaker(1)
}

func TestName_ReturnsNonEmpty(t *testing.T) {
	f := newTestFaker()
	name := f.Name()
	assert.NotEmpty(t, name)
	assert.Contains(t, name, " ", "Name should contain a space between first and last")
}

func TestFirstName_ReturnsNonEmpty(t *testing.T) {
	f := newTestFaker()
	assert.NotEmpty(t, f.FirstName())
}

func TestLastName_ReturnsNonEmpty(t *testing.T) {
	f := newTestFaker()
	assert.NotEmpty(t, f.LastName())
}

func TestEmail_ReturnsValidFormat(t *testing.T) {
	f := newTestFaker()
	email := f.Email()
	assert.NotEmpty(t, email)
	assert.Contains(t, email, "@")
	assert.Contains(t, email, ".")
}

func TestPhone_ReturnsNonEmpty(t *testing.T) {
	f := newTestFaker()
	phone := f.Phone()
	assert.NotEmpty(t, phone)
	assert.True(t, strings.HasPrefix(phone, "+1-"))
}

func TestAddress_ReturnsNonEmpty(t *testing.T) {
	f := newTestFaker()
	assert.NotEmpty(t, f.Address())
}

func TestCity_ReturnsNonEmpty(t *testing.T) {
	f := newTestFaker()
	assert.NotEmpty(t, f.City())
}

func TestCountry_ReturnsNonEmpty(t *testing.T) {
	f := newTestFaker()
	assert.NotEmpty(t, f.Country())
}

func TestUUID_ReturnsValidFormat(t *testing.T) {
	f := newTestFaker()
	uuid := f.UUID()
	assert.NotEmpty(t, uuid)
	// UUID v4 format: 8-4-4-4-12 hex chars
	parts := strings.Split(uuid, "-")
	assert.Len(t, parts, 5)
	assert.Len(t, parts[0], 8)
	assert.Len(t, parts[1], 4)
	assert.Len(t, parts[2], 4)
	assert.Len(t, parts[3], 4)
	assert.Len(t, parts[4], 12)
}

func TestParagraph_ReturnsNonEmpty(t *testing.T) {
	f := newTestFaker()
	p := f.Paragraph()
	assert.NotEmpty(t, p)
	assert.Contains(t, p, ".")
}

func TestSentence_ReturnsNonEmpty(t *testing.T) {
	f := newTestFaker()
	s := f.Sentence()
	assert.NotEmpty(t, s)
	assert.True(t, strings.HasSuffix(s, "."))
	// First character should be uppercase
	assert.Equal(t, strings.ToUpper(s[:1]), s[:1])
}

func TestWord_ReturnsNonEmpty(t *testing.T) {
	f := newTestFaker()
	assert.NotEmpty(t, f.Word())
}

func TestIntBetween_ReturnsValueInRange(t *testing.T) {
	f := newTestFaker()
	for i := 0; i < 100; i++ {
		v := f.IntBetween(5, 15)
		assert.GreaterOrEqual(t, v, 5)
		assert.LessOrEqual(t, v, 15)
	}
}

func TestIntBetween_MinEqualsMax(t *testing.T) {
	f := newTestFaker()
	v := f.IntBetween(7, 7)
	assert.Equal(t, 7, v)
}

func TestIntBetween_MinGreaterThanMax(t *testing.T) {
	f := newTestFaker()
	v := f.IntBetween(10, 5)
	assert.Equal(t, 10, v)
}

func TestFloat64Between_ReturnsValueInRange(t *testing.T) {
	f := newTestFaker()
	for i := 0; i < 100; i++ {
		v := f.Float64Between(1.5, 9.5)
		assert.GreaterOrEqual(t, v, 1.5)
		assert.Less(t, v, 9.5)
	}
}

func TestFloat64Between_MinEqualsMax(t *testing.T) {
	f := newTestFaker()
	v := f.Float64Between(3.14, 3.14)
	assert.Equal(t, 3.14, v)
}

func TestBool_ReturnsBoolean(t *testing.T) {
	f := newTestFaker()
	// Run enough times to see both values
	seenTrue := false
	seenFalse := false
	for i := 0; i < 100; i++ {
		if f.Bool() {
			seenTrue = true
		} else {
			seenFalse = true
		}
	}
	assert.True(t, seenTrue, "should produce true at least once")
	assert.True(t, seenFalse, "should produce false at least once")
}

func TestDate_ReturnsNonZero(t *testing.T) {
	f := newTestFaker()
	d := f.Date()
	assert.False(t, d.IsZero())
}

func TestDateBetween_ReturnsDateInRange(t *testing.T) {
	f := newTestFaker()
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	for i := 0; i < 100; i++ {
		d := f.DateBetween(start, end)
		assert.False(t, d.Before(start), "date should not be before start")
		assert.False(t, d.After(end), "date should not be after end")
	}
}

func TestDateBetween_StartEqualsEnd(t *testing.T) {
	f := newTestFaker()
	t0 := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
	d := f.DateBetween(t0, t0)
	assert.Equal(t, t0, d)
}

func TestPick_ReturnsElementFromSlice(t *testing.T) {
	f := newTestFaker()
	items := []string{"apple", "banana", "cherry"}
	for i := 0; i < 50; i++ {
		v := f.Pick(items)
		assert.Contains(t, items, v)
	}
}

func TestPick_EmptySlice(t *testing.T) {
	f := newTestFaker()
	v := f.Pick([]string{})
	assert.Empty(t, v)
}

func TestReproducibility(t *testing.T) {
	f1 := NewFaker(123)
	f2 := NewFaker(123)
	// Same seed should produce same sequence for deterministic methods
	assert.Equal(t, f1.FirstName(), f2.FirstName())
	assert.Equal(t, f1.LastName(), f2.LastName())
	assert.Equal(t, f1.IntBetween(0, 100), f2.IntBetween(0, 100))
	assert.Equal(t, f1.Word(), f2.Word())
}
