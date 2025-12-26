package auth

import (
	"context"
	"log/slog"
	domain "metrika/internal/domain/auth"
)

type RegisterUseCase struct {
	users    domain.UserRepository
	sessions domain.SessionRepository
	tokens   TokenProvider
	logger *slog.Logger
}

func NewRegisterUseCase(
	users domain.UserRepository,
	sessions domain.SessionRepository,
	tokens TokenProvider,
	logger *slog.Logger,
) *RegisterUseCase {
	return &RegisterUseCase{users, sessions, tokens, logger}
}

func (uc *RegisterUseCase) Execute(
	ctx context.Context,
	email string,
	passwordRaw string,
	passwordSecondRaw string,
	userAgent string,
) (*domain.Tokens, error) {

	//проверяем валидность паролей
	if !domain.ComparePasswords(passwordRaw, passwordSecondRaw){
		return nil, domain.ErrPasswordNotCompare
	}

	//хэшируем пароль
	hash, err := domain.HashPassword(passwordRaw)
	if err != nil {
		return nil, domain.ErrInvalidPasswordRaw
	}

	password := domain.NewPasswordFromHash(hash)

	auser := domain.User{
		Email:    email,
		Password: password,
	}

	//создаем юзера
	if err := uc.users.CreateUser(ctx, &auser); err != nil {
		return nil, err
	}


	//создаем сессию для юзера
	session := &domain.Session{
		UserID:    auser.ID,
		UserAgent: userAgent,
	}

	if err := uc.sessions.Create(ctx, session); err != nil {
		return nil, err
	}

	//генерируем рефреш токен для юзера
	tokens, err := uc.tokens.GeneratePair(
		email,
		session.UserID,
		session.ID,
	)
	if err != nil {
		return nil, err
	}

	session.RefreshToken = tokens.Refresh

	if err := uc.sessions.Update(ctx, session); err != nil {
		return nil, err
	}

	return tokens, nil
}
