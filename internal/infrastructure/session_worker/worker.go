package sessionworker

import (
	"log/slog"
	"time"
)

type SessionsWorker struct {
	log      *slog.Logger
	interval time.Duration
	fn       SessionsWorkerAdapter
}

type SessionsWorkerAdapter interface {
	cleanupBatchSessions(limit int)
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
			s.fn.cleanupBatchSessions(10)
			// if err := s.cleanupBatchSessions(10); err != nil {
			// logger.Error("ошибка при закрытии устаревших сессий", sl.Err(err))
			// }
		}()
	}
}
