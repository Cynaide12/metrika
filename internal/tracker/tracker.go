package tracker

import (
	"metrika/internal/models"
	"time"
)

type Tracker struct {
	Events    chan models.Event
	BatchSize int
	BuferSize int64
	Interval  time.Duration
	handler   StorageHandler
}

type StorageHandler interface {
	SaveEvents(events []models.Event)
}

func New(batchSize int, interval time.Duration, buferSize int64, handler StorageHandler) *Tracker {
	tr := Tracker{
		Events:    make(chan models.Event, buferSize),
		Interval:  interval,
		BuferSize: buferSize,
		BatchSize: batchSize,
		handler:   handler,
	}

	return &tr
}

func (r *Tracker) saver() {
	ticker := time.NewTicker(r.Interval)
	batch := make([]models.Event, 0, r.BatchSize)

	for {
		select {
		case e := <-r.Events:
			batch = append(batch, e)

			if len(batch) >= r.BatchSize {
				r.handler.SaveEvents(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			r.handler.SaveEvents(batch)
			batch = batch[:0]
		}
	}
}

func (r *Tracker) TrackEvent(e models.Event) {
	select {
	case r.Events <- e:

	default:

	}
}
