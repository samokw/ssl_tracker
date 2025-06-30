package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/samokw/ssl_tracker/internal/ssl"
	"github.com/samokw/ssl_tracker/internal/types"
)

type Service struct {
	domainRepo *Repository
	sslService *ssl.CertService
}

func NewService(domainRepo *Repository, sslService *ssl.CertService) *Service {
	return &Service{
		domainRepo: domainRepo,
		sslService: sslService,
	}
}

func (s *Service) AddDomain(userID types.UserID, domainName string) (*Domain, error) {
	err := ssl.ValidateHostnameDNS(domainName)
	if err != nil {
		return nil, err
	}
	domain := Domain{
		UserID:     userID,
		DomainName: NewDomainName(domainName),
		CreatedAt:  NewCreatedAt(time.Now()),
		IsActive:   true,
	}
	err = s.domainRepo.CreateDomain(&domain)
	if err != nil {
		return nil, err
	}

	hostname, err := ssl.NewHostname(domainName)
	if err != nil {
		return nil, fmt.Errorf("invalid hostname: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cert, err := ssl.CheckSSLCertificate(ctx, hostname)
	if err != nil {
		errorStr := err.Error()
		s.domainRepo.UpdateSSLInfo(domain.DomainID, nil, &errorStr)
	} else {
		expiryTime := cert.ExpiryDate.Time()
		s.domainRepo.UpdateSSLInfo(domain.DomainID, &expiryTime, nil)
	}

	return &domain, nil
}

func (s *Service) GetUsersDomains(userID types.UserID) ([]Domain, error) {
	return s.domainRepo.GetDomainsByUserID(userID)
}

func (s *Service) RemoveDomain(domainID types.DomainID) error {
	return s.domainRepo.DeleteDomain(domainID)
}

// CheckDomainSSL checks the SSL certificate for a specific domain
func (s *Service) CheckDomainSSL(domainID types.DomainID) error {
	// Get the domain from database
	domain, err := s.domainRepo.GetDomainByID(domainID)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	// Validate and create hostname
	hostname, err := ssl.NewHostname(domain.DomainName.String())
	if err != nil {
		// Update with error
		errorStr := err.Error()
		return s.domainRepo.UpdateSSLInfo(domainID, nil, &errorStr)
	}

	// Check SSL certificate
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cert, err := ssl.CheckSSLCertificate(ctx, hostname)
	if err != nil {
		// Update with error
		errorStr := err.Error()
		return s.domainRepo.UpdateSSLInfo(domainID, nil, &errorStr)
	}

	// Update with successful result
	expiryTime := cert.ExpiryDate.Time()
	return s.domainRepo.UpdateSSLInfo(domainID, &expiryTime, nil)
}

// CheckAllDomainsSSLSync checks SSL certificates for all domains synchronously and waits for completion
func (s *Service) CheckAllDomainsSSLSync(userID types.UserID) error {
	domains, err := s.GetUsersDomains(userID)
	if err != nil {
		return fmt.Errorf("failed to get domains: %w", err)
	}

	if len(domains) == 0 {
		return nil
	}

	// Use a channel to track completion
	done := make(chan bool, len(domains))

	// Start the SSL service (now safe to call multiple times)
	s.sslService.Start()

	// Set up result handler to update the database and signal completion
	s.sslService.SetResultHandler(func(result ssl.Result) {
		if result.Error != nil {
			errorStr := result.Error.Error()
			s.domainRepo.UpdateSSLInfo(types.DomainID(result.Task.DomainID), nil, &errorStr)
		} else {
			expiryTime := result.Certificate.ExpiryDate.Time()
			s.domainRepo.UpdateSSLInfo(types.DomainID(result.Task.DomainID), &expiryTime, nil)
		}
		done <- true
	})

	// Submit all domains to the worker pool
	for _, domain := range domains {
		s.sslService.CheckDomain(
			domain.DomainName.String(),
			int(domain.DomainID),
			int(userID),
		)
	}

	// Wait for all domains to be processed
	for i := 0; i < len(domains); i++ {
		<-done
	}

	return nil
}
