package auth

import (
    "context"
    domain "metrika/internal/domain/auth"
)

type RefreshUseCase struct {
    sessions domain.SessionRepository
    tokens   TokenProvider
}

func (uc *RefreshUseCase) Execute(
    ctx context.Context,
    refreshToken string,
) (domain.Tokens, error) {

    claims, err := uc.tokens.Validate(refreshToken)
    if err != nil {
        return domain.Tokens{}, domain.ErrInvalidRefreshToken
    }

    session, err := uc.sessions.ByID(ctx, claims.SessionID)
    if err != nil {
        return domain.Tokens{}, err
    }

    if session.RefreshToken != refreshToken {
        return domain.Tokens{}, domain.ErrInvalidRefreshToken
    }

    tokens, err := uc.tokens.GeneratePair(
		claims.Email,
        session.UserID,
        session.ID,
    )
    if err != nil {
        return domain.Tokens{}, err
    }

    session.RefreshToken = tokens.Refresh

    if err := uc.sessions.Update(ctx, session); err != nil {
        return domain.Tokens{}, err
    }

    return *tokens, nil
}
