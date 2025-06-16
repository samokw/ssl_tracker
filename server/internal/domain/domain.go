package domain

import (
	"time"

	"github.com/samokw/ssl_tracker/internal/types"
)

type DomainName string

type CreatedAt time.Time

type ExpiryDate time.Time

type Domain struct {
	DomainID   types.DomainID
	UserID     types.UserID
	DomainName DomainName
	CreatedAt  CreatedAt
	ExpiryDate ExpiryDate
}
