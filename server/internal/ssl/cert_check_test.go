package ssl

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateHostname_Valid - these hostnames should pass validation.
func TestValidateHostname_Valid(t *testing.T) {
	valid := []string{
		"example.com",
		"sub.example.com",
		"my-domain.com",
		"example123.com",
	}

	for _, h := range valid {
		t.Run(h, func(t *testing.T) {
			err := ValidateHostname(h)
			assert.NoError(t, err)
		})
	}
}

// TestValidateHostname_Invalid - these should fail validation.
func TestValidateHostname_Invalid(t *testing.T) {
	tests := []struct {
		hostname string
		wantErr  error
	}{
		{"", ErrEmptyHostname},
		{"   ", ErrEmptyHostname},
		{"example..com", ErrInvalidCharacters},
		{"-example.com", ErrInvalidCharacters},
		{"example.com:443", ErrInvalidCharacters},
		{strings.Repeat("a", 254), ErrHostnameTooLong},
	}

	for _, tc := range tests {
		t.Run(tc.hostname, func(t *testing.T) {
			err := ValidateHostname(tc.hostname)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

// TestNewHostname - creates a validated hostname.
func TestNewHostname(t *testing.T) {
	// Valid hostname
	h, err := NewHostname("example.com")
	require.NoError(t, err)
	assert.Equal(t, "example.com", h.String())
	assert.True(t, h.IsValid())

	// Invalid hostname
	h, err = NewHostname("")
	require.Error(t, err)
	assert.Equal(t, Hostname(""), h)
}

// TestCheckSSLCertificate_CancelledContext - returns error if context cancelled.
func TestCheckSSLCertificate_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	hostname, _ := NewHostname("example.com")
	_, err := CheckSSLCertificate(ctx, hostname)

	assert.Error(t, err)
}

// TestCheckSSLCertificate_InvalidHostname - returns error for empty hostname.
func TestCheckSSLCertificate_InvalidHostname(t *testing.T) {
	ctx := context.Background()
	hostname := Hostname("") // Empty = invalid

	_, err := CheckSSLCertificate(ctx, hostname)

	assert.ErrorIs(t, err, ErrInvalidHostname)
}

// TestCheckSSLCertificate_RealConnection - actually connects to a server.
// Skipped in short mode because it needs network.
func TestCheckSSLCertificate_RealConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	hostname, err := NewHostname("google.com")
	require.NoError(t, err)

	cert, err := CheckSSLCertificate(ctx, hostname)
	require.NoError(t, err)

	assert.Equal(t, hostname, cert.Hostname)
	assert.Greater(t, int(cert.TimeLeft), 0) // Should have days left
}

// FuzzValidateHostname - throws random strings at validation to find crashes.
func FuzzValidateHostname(f *testing.F) {
	// Seed with some examples
	f.Add("example.com")
	f.Add("")
	f.Add("a")
	f.Add(strings.Repeat("x", 300))

	f.Fuzz(func(t *testing.T, hostname string) {
		// Should never panic, just return nil or an error
		_ = ValidateHostname(hostname)
	})
}

// FuzzNewHostname - throws random strings at the constructor.
func FuzzNewHostname(f *testing.F) {
	f.Add("example.com")
	f.Add("")

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic
		_, _ = NewHostname(input)
	})
}
