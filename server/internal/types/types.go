package types

import (
	"fmt"
	"time"
)

type UserID uint

type ExpiryDate time.Time

type DomainID uint

// UserID helper functions
func NewUserID(id uint) UserID {
	return UserID(id)
}

func (u UserID) Uint() uint {
	return uint(u)
}

func ValidateUserID(userID UserID) error {
	if userID == 0 {
		return fmt.Errorf("user ID cannot be zero")
	}
	return nil
}

// DomainID helper functions
func NewDomainID(id uint) DomainID {
	return DomainID(id)
}

func (d DomainID) Uint() uint {
	return uint(d)
}

func ValidateDomainID(domainID DomainID) error {
	if domainID == 0 {
		return fmt.Errorf("domain ID cannot be zero")
	}
	return nil
}

func NewExpiryDate(t time.Time) ExpiryDate {
	return ExpiryDate(t)
}

func (e ExpiryDate) Time() time.Time {
	return time.Time(e)
}

func (e ExpiryDate) String() string {
	return time.Time(e).Format(time.RFC3339)
}
