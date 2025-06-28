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
