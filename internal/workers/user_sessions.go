package worker

import (
	"log/slog"
	"metrika/internal/repository"
	"time"
)

type SessionsWorker struct {
	log            *slog.Logger
	repo           *repository.Repository
	interval time.Duration
}

func NewSessionsWorker(log *slog.Logger, repo *repository.Repository, interval time.Duration) *SessionsWorker {
	return &SessionsWorker{
		log,
		repo,
		interval,
	}
}

func (s SessionsWorker) StartSessionManager(){
	ticker := time.NewTicker(s.interval)

	for range ticker.C{
		
	}
}