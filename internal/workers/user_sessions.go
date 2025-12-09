package worker

import (
	"errors"
	"log/slog"
	"metrika/internal/repository"
	"metrika/lib/logger/sl"
	"time"
)

type SessionsWorker struct {
	log      *slog.Logger
	repo     *repository.Repository
	interval time.Duration
}

func NewSessionsWorker(log *slog.Logger, repo *repository.Repository, interval time.Duration) *SessionsWorker {
	return &SessionsWorker{
		log,
		repo,
		interval,
	}
}

func (s SessionsWorker) StartSessionManager() {
	fn := "internal.workers.user_sessions.StartSessionManager"
	logger := s.log.With("fn", fn)
	ticker := time.NewTicker(s.interval)

	for range ticker.C {
		s.log.Debug("НАЧИНАЮ ЧИСТИТЬ СЕССИИ")
		go func() {
			if err := s.cleanupBatchSessions(10); err != nil {
				logger.Error("ошибка при закрытии устаревших сессий", sl.Err(err))
			}
		}()
	}
}

func (s SessionsWorker) cleanupBatchSessions(limit int) error {

	tx := s.repo.GormDB.Begin()

	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
		}
	}()

	txStorage := s.repo.WithTx(tx)

	sessions, err := txStorage.GetStaleSessions(limit)
	if errors.Is(err, repository.ErrNoRows) {
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

	s.log.Debug("ПОЧИЩЕНО СЕССИЙ", slog.Int("колво", int(len(session_ids))))

	return tx.Commit().Error
}
