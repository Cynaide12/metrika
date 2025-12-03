package mock

import (
	crypto "crypto/rand"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"metrika/internal/models"
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
}

func New(interval time.Duration, randWindow int, log *slog.Logger, bufferSize, maxEventInLoop, minEventInLoop int, tracker *tracker.Tracker) *Mock {
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

func (m *Mock) Stop() {
	close(m.closeChan)
}
