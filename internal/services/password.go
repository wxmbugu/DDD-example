package services

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytepassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytepassword), err
}

func CheckPassword(hashedpassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedpassword), []byte(password))
	return err
}
