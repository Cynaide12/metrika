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
	stop     chan struct{}
}

type SessionsWorkerAdapter interface {
	CleanupBatchSessions(ctx context.Context, limit int) error
}

func NewSessionsWorker(log *slog.Logger, interval time.Duration, fn SessionsWorkerAdapter, stop chan struct{}) *SessionsWorker {
	return &SessionsWorker{
		log,
		interval,
		fn,
		stop,
	}
}

func (s SessionsWorker) StartSessionManager() {
	ticker := time.NewTicker(s.interval)
	for c := ticker.C; ; {
		select {
		case <-c:
			go func() {
				ctx := context.Background()
				if err := s.fn.CleanupBatchSessions(ctx, 1000); err != nil {
					s.log.ErrorContext(ctx, "ошибка при закрытии неактивных сессий", sl.Err(err))
				}
			}()
		case <-s.stop:
			return
		}

	}
}
