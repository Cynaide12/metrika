package analytics

import (
	"context"
	"log/slog"
	domain "metrika/internal/domain/analytics"
	"metrika/internal/domain/tx"
)

type CleanupBatchSessionsUseCase struct {
	logger   *slog.Logger
	sessions domain.GuestSessionRepository
	tx tx.TransactionManager
}

func NewCleanupBatchSessionsUseCase(logger *slog.Logger, sessions domain.GuestSessionRepository, tx tx.TransactionManager) *CleanupBatchSessionsUseCase {
	return &CleanupBatchSessionsUseCase{logger, sessions, tx}
}

//TODO: доделать
func (c *CleanupBatchSessionsUseCase) CleanupBatchSessions(ctx context.Context,  limit int) error {


	sessions, err := c.sessions.GetStaleSessions(ctx, limit)
	if errors.Is(err, c.sessions.) {
		tx.Rollback()
		return nil
	}
	if err != nil {
		tx.Rollback()
		return err
	}

	//собираем id сессий
	var session_ids []uint
	for _, session := range sessions {
		session_ids = append(session_ids, session.ID)
	}

	//закрываем их
	if err := txStorage.CloseSession(session_ids); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
