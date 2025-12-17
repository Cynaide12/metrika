package auth

import (
	"metrika/pkg/password"
	"strings"
)

type Password struct {
	Hash []byte
}

func NewPasswordFromHash(hash []byte) Password {
	return Password{Hash: hash}
}

func HashPassword(raw string) ([]byte, error){
	return password.HashPassword(raw)
}

func ComparePasswords(password, second_password string) bool {
	return strings.Compare(password, second_password) == 0
}

func (p Password) Matches(raw string) bool {
	return password.CheckPasswordHash(raw, p.Hash)
}

func (p Password) ValidatePassword(password, second_password string) bool{
	return strings.Compare(password, second_password) == 0
}
