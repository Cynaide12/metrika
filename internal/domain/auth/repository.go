package auth

import (
	"context"
)

type UserRepository interface {
	ByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, auser *User) error
}

type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	ByID(ctx context.Context, id uint) (*Session, error)
	Update(ctx context.Context, s *Session) error
	Delete(ctx context.Context, id uint) error
}
