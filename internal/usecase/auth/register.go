package auth

import (
	"context"
	"log/slog"
	domain "metrika/internal/domain/auth"
	"metrika/internal/domain/tx"
)

type RegisterUseCase struct {
	users    domain.UserRepository
	sessions domain.SessionRepository
	tokens   TokenProvider
	logger   *slog.Logger
	tx       tx.TransactionManager
}

func NewRegisterUseCase(
	users domain.UserRepository,
	sessions domain.SessionRepository,
	tokens TokenProvider,
	logger *slog.Logger,
	tx tx.TransactionManager,
) *RegisterUseCase {
	return &RegisterUseCase{users, sessions, tokens, logger, tx}
}

func (uc *RegisterUseCase) Execute(
	ctx context.Context,
	email string,
	passwordRaw string,
	passwordSecondRaw string,
	userAgent string,
) (*domain.Tokens, error) {
	var tokens *domain.Tokens
	if err := uc.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		//проверяем валидность паролей
		if !domain.ComparePasswords(passwordRaw, passwordSecondRaw) {
			return domain.ErrPasswordNotCompare
		}

		//хэшируем пароль
		hash, err := domain.HashPassword(passwordRaw)
		if err != nil {
			return domain.ErrInvalidPasswordRaw
		}

		password := domain.NewPasswordFromHash(hash)

		auser := domain.User{
			Email:    email,
			Password: password,
		}

		//создаем юзера
		if err := uc.users.CreateUser(ctx, &auser); err != nil {
			return err
		}

		//создаем сессию для юзера
		session := &domain.Session{
			UserID:    auser.ID,
			UserAgent: userAgent,
		}

		if err := uc.sessions.Create(ctx, session); err != nil {
			return err
		}

		//генерируем рефреш токен для юзера
		tokens, err = uc.tokens.GeneratePair(
			email,
			session.UserID,
			session.ID,
		)
		if err != nil {
			return err
		}

		session.RefreshToken = tokens.Refresh

		if err := uc.sessions.Update(ctx, session); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return tokens, nil
}
