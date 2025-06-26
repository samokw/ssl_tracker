package user

import (
	"github.com/samokw/ssl_tracker/internal/types"
	"golang.org/x/crypto/bcrypt"
)

type Email string

type Password string

type WebHookUrl string

type User struct {
	UserID   types.UserID
	Email    Email
	Password Password
}

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (u *User) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
