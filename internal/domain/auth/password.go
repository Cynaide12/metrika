package auth

import "metrika/pkg/password"

type Password struct {
	hash []byte
}

func NewPasswordFromHash(hash []byte) Password {
	return Password{hash: hash}
}

func (p Password) Matches(raw string) bool {
	return password.CheckPasswordHash(raw, p.hash)
}
