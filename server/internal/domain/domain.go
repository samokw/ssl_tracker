package domain

import (
	"time"

	"github.com/samokw/ssl_tracker/internal/types"
)

type DomainName string
type CreatedAt time.Time
type ExpiryDate time.Time
type LastChecked time.Time
type LastError string // The type of error that occurred when checking

func NewDomainName(name string) DomainName {
	return DomainName(name)
}

func (d DomainName) String() string {
	return string(d)
}

func NewCreatedAt(t time.Time) CreatedAt {
	return CreatedAt(t)
}

func (c CreatedAt) Time() time.Time {
	return time.Time(c)
}

func (c CreatedAt) String() string {
	return time.Time(c).Format(time.RFC3339)
}

// ExpiryDate methods

func NewLastChecked(t time.Time) LastChecked {
	return LastChecked(t)
}

func (l LastChecked) Time() time.Time {
	return time.Time(l)
}

func (l LastChecked) String() string {
	return time.Time(l).Format(time.RFC3339)
}

func NewLastError(err string) LastError {
	return LastError(err)
}

func (l LastError) String() string {
	return string(l)
}

type Domain struct {
	DomainID    types.DomainID    `db:"id"`
	UserID      types.UserID      `db:"user_id"`
	DomainName  DomainName        `db:"domain_name"`
	CreatedAt   CreatedAt         `db:"created_at"`
	ExpiryDate  *types.ExpiryDate `db:"expiry_date"`
	LastChecked *LastChecked      `db:"last_checked"`
	LastError   *LastError        `db:"last_error"`
	IsActive    bool              `db:"is_active"`
}
