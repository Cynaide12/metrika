package analytics

import (
	"context"
	domain "metrika/internal/domain/analytics"
	"metrika/internal/domain/tx"
	"time"
)

type TrackerProvider interface {
	TrackEvent(e domain.Event)
}

type CollectEventsUseCase struct {
	events   domain.EventsRepository
	sessions domain.GuestSessionRepository
	tracker  TrackerProvider
	tx       tx.TransactionManager
}

func NewCollectEventsUseCase(
	events domain.EventsRepository,
	tracker TrackerProvider,
	sessions domain.GuestSessionRepository,
	tx tx.TransactionManager,
) *CollectEventsUseCase {
	return &CollectEventsUseCase{events, sessions, tracker, tx}
}

func (ec *CollectEventsUseCase) Execute(
	ctx context.Context,
	events *[]domain.Event,
) error {
	// ec.tracker.TrackEvent(event)
	return ec.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := ec.events.SaveEvents(ctx, events); err != nil {
			return err
		}

		var ids []uint
		for _, e := range *events {
			ids = append(ids, e.SessionID)
		}

		//TODO:учесть что с момента отправки задачи в очередь на сохранение может пройти много времени
		//TODO: соответственно time.Now может быть не актуален для данной задачи
		if err := ec.sessions.SetLastActive(ctx, ids, time.Now()); err != nil {
			return err
		}

		return nil
	})
}
