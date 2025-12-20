package tracker

import (
	"context"
	"log"
	domain "metrika/internal/domain/analytics"
	"time"
)

type Tracker struct {
	Events    chan domain.Event
	BatchSize int
	BuferSize int64
	Interval  time.Duration
	handler   StorageHandler
}

type StorageHandler interface {
	SaveEvents(ctx context.Context, events *[]domain.Event) error
}

func New(batchSize int, interval time.Duration, buferSize int64, handler StorageHandler) *Tracker {
	tr := Tracker{
		Events:    make(chan domain.Event, buferSize),
		Interval:  interval,
		BuferSize: buferSize,
		BatchSize: batchSize,
		handler:   handler,
	}

	go tr.saver()

	return &tr
}

func (r *Tracker) saver() {
	ticker := time.NewTicker(r.Interval)
	batch := make([]domain.Event, 0, r.BatchSize)

	ctx := context.Background()

	for {
		select {
		case e := <-r.Events:
			batch = append(batch, e)

			if len(batch) >= r.BatchSize {
				r.handler.SaveEvents(ctx, &batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				log.Println("СОХРАНЯЮ ИВЕНТ")
				r.handler.SaveEvents(ctx, &batch)
				batch = batch[:0]
			}
		}
	}
}

func (r *Tracker) TrackEvent(e domain.Event) {
	select {
	case r.Events <- e:

	default:

	}
}
