package analytics

import (
	"context"
	"errors"
	"log/slog"
	domain "metrika/internal/domain/analytics"
	"metrika/internal/domain/tx"
)

type CleanupBatchSessionsUseCase struct {
	logger   *slog.Logger
	sessions domain.GuestSessionRepository
	tx       tx.TransactionManager
}

func NewCleanupBatchSessionsUseCase(logger *slog.Logger, sessions domain.GuestSessionRepository, tx tx.TransactionManager) *CleanupBatchSessionsUseCase {
	return &CleanupBatchSessionsUseCase{logger, sessions, tx}
}

func (c *CleanupBatchSessionsUseCase) CleanupBatchSessions(ctx context.Context, limit int) error {

	sessions, err := c.sessions.GetStaleSessions(ctx, limit)
	if errors.Is(err, domain.ErrStaleSessionsNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	//собираем id сессий
	var session_ids []uint
	for _, session := range *sessions {
		session_ids = append(session_ids, session.ID)
	}

	//закрываем их
	if err := c.sessions.CloseSessions(ctx, session_ids); err != nil {
		return err
	}

	return nil
}
