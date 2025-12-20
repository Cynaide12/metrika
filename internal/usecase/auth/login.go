package auth

import (
	"context"
	domain "metrika/internal/domain/auth"
)

type TokenProvider interface {
	GeneratePair(email string, userID uint, session_id uint) (tokens *domain.Tokens, err error)
	Validate(tokenString string) (*domain.JWTClaims, error)
}

type LoginUseCase struct {
	users    domain.UserRepository
	sessions domain.SessionRepository
	tokens   TokenProvider
}

func NewLoginUseCase(
	users domain.UserRepository,
	sessions domain.SessionRepository,
	tokens TokenProvider,
) *LoginUseCase {
	return &LoginUseCase{users, sessions, tokens}
}

func (uc *LoginUseCase) Execute(
	ctx context.Context,
	email string,
	password string,
	userAgent string,
) (*domain.Tokens, error) {

	user, err := uc.users.ByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.Password.Matches(password) {
		return nil, domain.ErrInvalidCredentials
	}

	session := &domain.Session{
		UserID:    user.ID,
		UserAgent: userAgent,
	}

	if err := uc.sessions.Create(ctx, session); err != nil {
		return nil, err
	}

	tokens, err := uc.tokens.GeneratePair(user.Email, user.ID, session.ID)
	if err != nil {
		return nil, err
	}

	if err := uc.sessions.Update(ctx, session); err != nil {
		return nil, err
	}

	return tokens, nil
}
