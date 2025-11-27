package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestDomainName - basic creation and String() method.
func TestDomainName(t *testing.T) {
	dn := NewDomainName("example.com")

	assert.Equal(t, "example.com", dn.String())
	assert.Equal(t, DomainName("example.com"), dn)
}

// TestDomainName_Empty - empty string is allowed (validation happens elsewhere).
func TestDomainName_Empty(t *testing.T) {
	dn := NewDomainName("")
	assert.Equal(t, "", dn.String())
}

// TestCreatedAt - stores and formats time correctly.
func TestCreatedAt(t *testing.T) {
	now := time.Now()
	ca := NewCreatedAt(now)

	assert.Equal(t, now, ca.Time())
	assert.NotEmpty(t, ca.String())      // Should have some string
	assert.Contains(t, ca.String(), "T") // RFC3339 format has "T" in it
}

// TestLastChecked - stores time correctly.
func TestLastChecked(t *testing.T) {
	now := time.Now()
	lc := NewLastChecked(now)

	assert.Equal(t, now, lc.Time())
	assert.NotEmpty(t, lc.String())
}

// TestLastError - stores error message.
func TestLastError(t *testing.T) {
	le := NewLastError("connection timeout")
	assert.Equal(t, "connection timeout", le.String())

	// Empty error is also valid
	empty := NewLastError("")
	assert.Equal(t, "", empty.String())
}

// TestDomain_Struct - the full Domain struct.
func TestDomain_Struct(t *testing.T) {
	now := time.Now()

	domain := Domain{
		DomainName:  NewDomainName("example.com"),
		CreatedAt:   NewCreatedAt(now),
		LastChecked: nil, // optional
		LastError:   nil, // optional
		IsActive:    true,
	}

	assert.Equal(t, "example.com", domain.DomainName.String())
	assert.True(t, domain.IsActive)
	assert.Nil(t, domain.LastChecked)
	assert.Nil(t, domain.LastError)
}

// TestDomain_WithAllFields - Domain with all optional fields set.
func TestDomain_WithAllFields(t *testing.T) {
	now := time.Now()
	lastChecked := NewLastChecked(now.Add(-1 * time.Hour))
	lastError := NewLastError("previous error")

	domain := Domain{
		DomainName:  NewDomainName("test.example.com"),
		CreatedAt:   NewCreatedAt(now),
		LastChecked: &lastChecked,
		LastError:   &lastError,
		IsActive:    false,
	}

	assert.Equal(t, "test.example.com", domain.DomainName.String())
	assert.NotNil(t, domain.LastChecked)
	assert.NotNil(t, domain.LastError)
	assert.Equal(t, "previous error", domain.LastError.String())
	assert.False(t, domain.IsActive)
}

// FuzzDomainName - random strings shouldn't crash.
func FuzzDomainName(f *testing.F) {
	f.Add("example.com")
	f.Add("")
	f.Add(strings.Repeat("x", 1000))

	f.Fuzz(func(t *testing.T, input string) {
		dn := NewDomainName(input)
		assert.Equal(t, input, dn.String())
	})
}

// FuzzLastError - random error messages shouldn't crash.
func FuzzLastError(f *testing.F) {
	f.Add("")
	f.Add("error message")
	f.Add(strings.Repeat("x", 1000))

	f.Fuzz(func(t *testing.T, input string) {
		le := NewLastError(input)
		assert.Equal(t, input, le.String())
	})
}
