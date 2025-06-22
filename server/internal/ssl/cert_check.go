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

type TimeLeft int
type Hostname string
type ExpiryDate time.Time

type SLLCertificate struct {
	Hostname   Hostname
	ExpiryDate ExpiryDate
	TimeLeft   TimeLeft
}

var (
	InvalidHostnameErr   = errors.New("invalid hostname")
	HostnameTooLongErr   = errors.New("hostname too long")
	InvalidCharactersErr = errors.New("hostname contains invalid characters")
	EmptyHostnameErr     = errors.New("hostname cannot be empty")
)

func ValidateHostname(hostname string) error {
	if strings.TrimSpace(hostname) == "" {
		return EmptyHostnameErr
	}

	if len(hostname) > 253 {
		return HostnameTooLongErr
	}
	validHostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !validHostnameRegex.MatchString(hostname) {
		return InvalidCharactersErr
	}

	if strings.Contains(hostname, "..") {
		return InvalidHostnameErr
	}
	if strings.HasPrefix(hostname, ".") || strings.HasSuffix(hostname, ".") ||
		strings.HasPrefix(hostname, "-") || strings.HasSuffix(hostname, "-") {
		return InvalidHostnameErr
	}
	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return InvalidHostnameErr
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return InvalidHostnameErr
		}
	}
	return nil
}

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

func NewHostname(hostname string) (Hostname, error) {
	if err := ValidateHostname(hostname); err != nil {
		return "", err
	}
	return Hostname(hostname), nil
}

func (h Hostname) String() string {
	return string(h)
}

func (h Hostname) IsValid() bool {
	return ValidateHostname(h.String()) == nil
}

func CheckSSLCertificate(ctx context.Context ,hostname Hostname) (*SLLCertificate, error) {
	logger := slog.With("hostname", hostname.String(), "operation", "ssl_check")
	if !hostname.IsValid() {
		logger.Error("Invalid hostname provided")
		return nil, InvalidHostnameErr
	}

	logger.Info("Starting SSL certificate check")
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(hostname.String(), "443"), 10*time.Second)
	if err != nil {
		logger.Error("Failed to establish TCP connection", "error", err)
		return nil, fmt.Errorf("failed to connect to %s: %w", hostname, err)
	}
	defer conn.Close()

	logger.Debug("TCP connection established")

	client := tls.Client(conn, &tls.Config{
		ServerName: hostname.String(),
	})

	err = client.Handshake()
	if err != nil {
		logger.Error("TLS handshake failed", "error", err)
		return nil, fmt.Errorf("TLS handshake failed for %s: %w", hostname, err)
	}
	defer client.Close()

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
	
	return &SLLCertificate{
		Hostname:   hostname,
		ExpiryDate: expiryDate,
		TimeLeft:   timeLeft,
	}, nil
}
