package domain

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/samokw/ssl_tracker/internal/types"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) scanDomainRow(row *sql.Row) (Domain, error) {
	// We need to use default types and then convert them to our types
	var domainID, userID uint
	var domainName string
	var createdAt time.Time
	var expiryDate, lastChecked sql.NullTime
	var lastError sql.NullString
	var isActive bool

	// scan information from the database
	err := row.Scan(&domainID, &userID, &domainName, &createdAt, &expiryDate, &lastChecked, &lastError, &isActive)
	if err != nil {
		return Domain{}, err
	}

	// Create the object domain we will return
	domain := Domain{
		DomainID:   types.DomainID(domainID),
		UserID:     types.UserID(userID),
		DomainName: NewDomainName(domainName),
		CreatedAt:  NewCreatedAt(createdAt),
		IsActive:   isActive,
	}
	if expiryDate.Valid {
		ed := types.NewExpiryDate(expiryDate.Time)
		domain.ExpiryDate = &ed
	} else {
		domain.ExpiryDate = nil
	}
	if lastChecked.Valid {
		lc := NewLastChecked(lastChecked.Time)
		domain.LastChecked = &lc
	} else {
		domain.LastChecked = nil
	}
	if lastError.Valid {
		le := NewLastError(lastError.String)
		domain.LastError = &le
	} else {
		domain.LastError = nil
	}
	return domain, nil
}

func (r *Repository) scanDomain(rows *sql.Rows) (Domain, error) {
	// We need to use default types and then convert them to our types
	var domainID, userID uint
	var domainName string
	var createdAt time.Time
	var expiryDate, lastChecked sql.NullTime
	var lastError sql.NullString
	var isActive bool

	// scan information from the database
	err := rows.Scan(&domainID, &userID, &domainName, &createdAt, &expiryDate, &lastChecked, &lastError, &isActive)
	if err != nil {
		return Domain{}, err
	}

	// Create the object domain we will return
	domain := Domain{
		DomainID:   types.DomainID(domainID),
		UserID:     types.UserID(userID),
		DomainName: NewDomainName(domainName),
		CreatedAt:  NewCreatedAt(createdAt),
		IsActive:   isActive,
	}
	if expiryDate.Valid {
		ed := types.NewExpiryDate(expiryDate.Time)
		domain.ExpiryDate = &ed
	} else {
		domain.ExpiryDate = nil
	}
	if lastChecked.Valid {
		lc := NewLastChecked(lastChecked.Time)
		domain.LastChecked = &lc
	} else {
		domain.LastChecked = nil
	}
	if lastError.Valid {
		le := NewLastError(lastError.String)
		domain.LastError = &le
	} else {
		domain.LastError = nil
	}
	return domain, nil
}

func (r *Repository) CheckForDuplicateDomains(userID types.UserID, domainName string) (*Domain, error) {
	query := `SELECT id, user_id, domain_name, created_at, expiry_date, last_checked, last_error, is_active 
              FROM domains WHERE user_id = ? AND domain_name = ?`
	row := r.db.QueryRow(query, userID.Uint(), domainName)
	domain, err := r.scanDomainRow(row)
	if err != nil {
		if err == sql.ErrNoRows { // We found no duplicate
			return nil, nil
		}
		return nil, err
	}
	// This is the duplicate domain we found
	return &domain, nil
}

func (r *Repository) CreateDomain(domain *Domain) error {
	if err := types.ValidateUserID(domain.UserID); err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	if domain.DomainName.String() == "" {
		return fmt.Errorf("domain name cannot be empty")
	}
	existingDomain, err := r.CheckForDuplicateDomains(domain.UserID, domain.DomainName.String())
	if err != nil {
		return fmt.Errorf("error checking for duplicate domain: %w", err)
	}
	if existingDomain != nil {
		return fmt.Errorf("domain %s already exists for this user", domain.DomainName.String())
	}
	query := `INSERT INTO domains (user_id, domain_name, is_active, created_at) VALUES (?, ?, ?, ?)`
	result, err := r.db.Exec(query, domain.UserID.Uint(), domain.DomainName.String(), domain.IsActive, domain.CreatedAt.Time())
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	domain.DomainID = types.NewDomainID(uint(id))
	return err
}

func (r *Repository) GetDomainsByUserID(userID types.UserID) ([]Domain, error) {
	query := `SELECT id, user_id, domain_name, created_at, expiry_date, last_checked, last_error, is_active FROM domains WHERE user_id = ?`
	rows, err := r.db.Query(query, userID.Uint())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	domains := []Domain{}

	for rows.Next() {
		domain, err := r.scanDomain(rows)
		if err != nil {
			return nil, err
		}
		domains = append(domains, domain)
	}
	return domains, nil
}

// View a domain by its ID
func (r *Repository) GetDomainByID(domainID types.DomainID) (*Domain, error) {
	query := `SELECT id, user_id, domain_name, created_at, expiry_date, last_checked, last_error, is_active FROM domains WHERE id = ?`
	row := r.db.QueryRow(query, domainID.Uint())
	domain, err := r.scanDomainRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("domain with ID %d not found", domainID.Uint())
		}
		return nil, err
	}
	return &domain, err
}

// Delete A domain by its ID
func (r *Repository) DeleteDomain(domainID types.DomainID) error {
	query := `DELETE FROM domains WHERE id = ?`
	result, err := r.db.Exec(query, domainID.Uint())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("domain with ID %d not found", domainID.Uint())
	}

	return nil
}

// Update A domains info based on the ssl check
func (r *Repository) UpdateSSLInfo(domainID types.DomainID, expiryDate *time.Time, lastError *string) error {
	now := time.Now()
	query := `UPDATE domains SET expiry_date = ?, last_checked = ?, last_error = ? WHERE id = ?`

	var expiryNull sql.NullTime
	var errorNull sql.NullString

	if expiryDate != nil {
		expiryNull.Time = *expiryDate
		expiryNull.Valid = true
	} else {
		expiryNull.Valid = false
	}

	if lastError != nil {
		errorNull.String = *lastError
		errorNull.Valid = true
	} else {
		errorNull.Valid = false
	}
	result, err := r.db.Exec(query, expiryNull, now, errorNull, domainID.Uint())
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("domain with ID %d not found", domainID.Uint())
	}
	return nil
}
