// This package provides SSL certificate checking
//
// It includes a way to validate hostnames, and provides information on the expiry of the certificates
package ssl

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// TimeLeft represents the days left until an SSL certificate until it expires
type TimeLeft int

// Hostname represents a validated domain name
type Hostname string

// ExpiryDate represents the Date when the SSL certificate will expire
type ExpiryDate time.Time

// SSLCertificate represents SSL certificate infromation.
//
// This includes Hostname, Expiry Date, and Time Remaining until it expires
type SSLCertificate struct {
	// Hostname is the domain name this certificate is valid for
	Hostname Hostname
	// ExpiryDate is when the certificate expires
	ExpiryDate ExpiryDate
	// TimeLeft is the number days left until the certificate expires
	TimeLeft TimeLeft
}

// Common hostname validation errors.
var (
	// ErrInvalidHostname occurs when the hostname is invalid
	ErrInvalidHostname = errors.New("invalid hostname")
	// ErrHostnameTooLong occurs when the hostname is too long
	ErrHostnameTooLong = errors.New("hostname too long")
	// ErrInvalidCharacters occurs when a hostname has an invalid character
	ErrInvalidCharacters = errors.New("hostname contains invalid characters")
	// ErrEmptyHostname occurs when the hostname is empty
	ErrEmptyHostname = errors.New("hostname cannot be empty")
)

// ValidateHostname checks if a hostname string is valid
//
// The validation checks for:
//   - Empty hostnames
//   - Lengths of the domain
//   - Validity of characters
//   - Proper formatting
//
// Returns nil if the format is valid, or one the defined errors if a problem is found
func ValidateHostname(hostname string) error {
	if strings.TrimSpace(hostname) == "" {
		return ErrEmptyHostname
	}

	if len(hostname) > 253 {
		return ErrHostnameTooLong
	}
	validHostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !validHostnameRegex.MatchString(hostname) {
		return ErrInvalidCharacters
	}

	if strings.Contains(hostname, "..") {
		return ErrInvalidHostname
	}
	if strings.HasPrefix(hostname, ".") || strings.HasSuffix(hostname, ".") ||
		strings.HasPrefix(hostname, "-") || strings.HasSuffix(hostname, "-") {
		return ErrInvalidHostname
	}
	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return ErrInvalidHostname
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return ErrInvalidHostname
		}
	}
	return nil
}

// ValidateHostnameDNS checks if a hostname can be resolved
//
// # It first validates the format, then runs a Host lookup on the hostname
//
// Returns nil if the hostname is valid and is found, or an error if the validation or the hostnamne is not found
func ValidateHostnameDNS(hostname string) error {
	if err := ValidateHostname(hostname); err != nil {
		return err
	}
	_, err := net.LookupHost(hostname)
	if err != nil {
		return errors.New("could not find the hostname: " + err.Error())
	}
	return nil
}

// ValidateURL checks if a hostname is valid when used as part of a URL.
//
// This function validates that the hostname can be parsed as part of
// a valid HTTPS URL structure and passes hostname validation.
//
// Returns nil if the hostname is valid for URL use, or an error if invalid.
func ValidateURL(hostname string) error {
	urlStruct, err := url.Parse("https://" + hostname)
	if err != nil {
		return errors.New("invalid URL format: " + err.Error())
	}
	if urlStruct.Host == "" {
		return errors.New("no host found in URL")
	}
	return ValidateHostname(urlStruct.Host)
}

// NewHostname creates a new Hostname after validating it
//
// # Recommended way to create a Hostname as it ensures the hostname is valid
//
// Returns the validated Hostname or an error if the validation fails
func NewHostname(hostname string) (Hostname, error) {
	if err := ValidateHostname(hostname); err != nil {
		return "", err
	}
	return Hostname(hostname), nil
}

// String returns the hostname as a string.
// This implements the fmt.Stringer interface.
func (h Hostname) String() string {
	return string(h)
}

// IsValid chceks if the hostname is still valid according to the rules
//
// Serves as a quick way to validate in code
func (h Hostname) IsValid() bool {
	return ValidateHostname(h.String()) == nil
}

// CheckSSLCertificate does a SSL certificate check on the provided hostname.
//
// 1. It Establishes a TCP connection on the HTTPS port (443)
// 2. Preforms a TCP handshake (SYN-SYN-ACK)
// 3. Retrieves the server's SSL certificate
// 4. Calculates the expiry Infomation
//
// Returns SSL certificate information or an error if a check failed
func CheckSSLCertificate(ctx context.Context, hostname Hostname) (*SSLCertificate, error) {
	logger := slog.With("hostname", hostname.String(), "operation", "ssl_check")
	if !hostname.IsValid() {
		logger.Error("Invalid hostname provided")
		return nil, ErrInvalidHostname
	}

	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}
	logger.Info("Starting SSL certificate check")
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(hostname.String(), "443"))
	if err != nil {
		logger.Error("Failed to establish TCP connection", "error", err)
		return nil, fmt.Errorf("failed to connect to %s: %w", hostname, err)
	}
	defer conn.Close()

	logger.Debug("TCP connection established")

	client := tls.Client(conn, &tls.Config{
		ServerName: hostname.String(),
	})
	err = client.HandshakeContext(ctx)
	if err != nil {
		logger.Error("TLS handshake failed", "error", err)
		return nil, fmt.Errorf("TLS handshake failed for %s: %w", hostname, err)
	}
	defer client.Close()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	logger.Debug("TLS handshake completed")
	certs := client.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		logger.Error("No certificates found")
		return nil, fmt.Errorf("no certificates found for %s", hostname)
	}

	cert := certs[0]
	expiryDate := ExpiryDate(cert.NotAfter)
	timeLeft := TimeLeft(time.Until(cert.NotAfter).Hours() / 24)

	logger.Info("SSL certificate check completed",
		"expires_at", cert.NotAfter,
		"days_remaining", int(timeLeft),
		"issuer", cert.Issuer.CommonName,
	)

	return &SSLCertificate{
		Hostname:   hostname,
		ExpiryDate: expiryDate,
		TimeLeft:   timeLeft,
	}, nil
}
