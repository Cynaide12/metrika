package mock

import (
	"fmt"
	"log/slog"
	"math/rand"
	"metrika/internal/models"
	"metrika/internal/tracker"
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
	}
}

func (m Mock) StartEventsGenerator() {
	ticker := time.NewTicker(m.interval)

	for {

		select {
		case <-ticker.C:
			var bucketSize int = rand.Intn(m.maxEventInLoop)
			for {
				if bucketSize < m.minEventInLoop {
					bucketSize = rand.Intn(m.maxEventInLoop)
				}
				break
			}
			log := m.log.With("bucketSize", bucketSize, "time", time.Now().String())
			log.Info("начал добавлять события")
			for i := range bucketSize {
				UserID := fmt.Sprintf("%d", time.Now().Nanosecond())
				SessionID := fmt.Sprintf("%d", time.Now().Nanosecond()+i)
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

func (m Mock) Stop(){
	close(m.closeChan)
}
