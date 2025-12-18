package sessionworker

import (
	"context"
	"log/slog"
	"metrika/pkg/logger/sl"
	"time"
)

type SessionsWorker struct {
	log      *slog.Logger
	interval time.Duration
	fn       SessionsWorkerAdapter
}

type SessionsWorkerAdapter interface {
	CleanupBatchSessions(ctx context.Context, limit int) error
}

func NewSessionsWorker(log *slog.Logger, interval time.Duration, fn SessionsWorkerAdapter) *SessionsWorker {
	return &SessionsWorker{
		log,
		interval,
		fn,
	}
}

func (s SessionsWorker) StartSessionManager() {
	ticker := time.NewTicker(s.interval)

	for range ticker.C {
		go func() {
			ctx := context.Background()
			if err := s.fn.CleanupBatchSessions(ctx, 10); err != nil {
				s.log.ErrorContext(ctx, "ошибка при закрытии неактивных сессий", sl.Err(err))
			}
			// if err := s.cleanupBatchSessions(10); err != nil {
			// logger.Error("ошибка при закрытии устаревших сессий", sl.Err(err))
			// }
		}()
	}
}
