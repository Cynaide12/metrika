package mock

import (
	crypto "crypto/rand"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"metrika/internal/models"
	"metrika/internal/repository"
	"metrika/internal/tracker"
	"sync/atomic"
	"time"
)

type Mock struct {
	interval       time.Duration
	randWindow     int
	log            *slog.Logger
	EventsChan     chan models.Event
	bufferSize     int
	maxEventInLoop int
	minEventInLoop int
	tracker        *tracker.Tracker
	closeChan      chan struct{}
	idsCounter     atomic.Int64
	mockService    MockService
}

type MockService interface {
	GetDomains(domains *[]models.Domain, opts repository.GetDomainsOptions) error
}

func New(interval time.Duration, randWindow int, log *slog.Logger, bufferSize, maxEventInLoop, minEventInLoop int, tracker *tracker.Tracker, repo repository.Repository) *Mock {
	return &Mock{
		interval,
		randWindow,
		log,
		make(chan models.Event, bufferSize),
		bufferSize,
		maxEventInLoop,
		minEventInLoop,
		tracker,
		make(chan struct{}),
		atomic.Int64{},
		repo,
	}
}

func (m *Mock) generateRandomUuid() string {
	var buf [16]byte
	//сначала записываем в слайс случайные байты
	_, err := crypto.Read(buf[:])
	//если генератор не работает - генерируем свой uuid
	if err != nil {
		t := time.Now().UnixMicro()
		c := m.idsCounter.Add(1)
		return fmt.Sprintf("%d-%d", t, c)
	}
	//преобразуем слайс байтов в шестнадцатеричную строку
	return fmt.Sprintf("%x", buf[:])
}

func (m *Mock) generateBucketSize(min, max int) int {
	return min + rand.IntN(max-min)
}

func (m *Mock) StartEventsGenerator() {
	ticker := time.NewTicker(m.interval)

	defer func() {
		if r := recover(); r != nil {
			m.log.Error("events generator error", slog.Any("err", r))
		}
	}()

	for {
		select {
		case <-ticker.C:
			bucketSize := m.generateBucketSize(m.minEventInLoop, m.maxEventInLoop)

			log := m.log.With("bucketSize", bucketSize, "time", time.Now().String())
			log.Info("начал добавлять события")
			for range bucketSize {
				UserID := m.generateRandomUuid()
				SessionID := m.generateRandomUuid()
				m.tracker.TrackEvent(models.Event{
					UserID:    UserID,
					SessionID: SessionID,
					Type:      "mock",
					PageURL:   "mock",
					Timestamp: time.Now(),
				})
			}
			log.Info("закончил добавлять события")
		case <-m.closeChan:
			return
		}
	}
}

// TODO: доделать, надо сюда передавать список доменов для которых генериться будут юзеры(думаю лучше для одного)
// TODO: сделать еще функцию генерации сессий, она тоже должна принимать массив id юзеров для которых генерим сессию
func (m *Mock) StartUsersGenerator() {
	ticker := time.NewTicker(m.interval)

	defer func() {
		if r := recover(); r != nil {
			m.log.Error("events generator error", slog.Any("err", r))
		}
	}()

	for {
		select {
		case <-ticker.C:
			user := models.User{}
		case <-m.closeChan:
			return
		}
	}
}

func (m *Mock) Stop() {
	close(m.closeChan)
}
