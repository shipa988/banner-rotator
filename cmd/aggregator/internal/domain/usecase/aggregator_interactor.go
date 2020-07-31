package usecase

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/shipa988/banner_rotator/internal/data/logger"
	"github.com/shipa988/banner_rotator/internal/domain/entities"
)

const (
	ErrProcessClickEvent = `can't register click event for banner id: "%v", page: "%v", slot id: "%v"`
	ErrProcessShowEvent  = `can't register show event for banner id: "%v", page: "%v", slot id: "%v"`
)

var _ Aggregator = (*AggregatorInteractor)(nil)

type AggregatorInteractor struct {
	actionRepo entities.ActionRepository
	queue      entities.EventQueue
	logger     logger.Logger
}

func NewAggregatorInteractor(repo entities.ActionRepository, queueBroker entities.EventQueue, logger logger.Logger) (*AggregatorInteractor, error) {
	return &AggregatorInteractor{
		actionRepo: repo,
		queue:      queueBroker,
		logger:     logger,
	}, nil
}

func (a *AggregatorInteractor) processEvent(ctx context.Context, wg *sync.WaitGroup, events <-chan entities.Event) {
	defer wg.Done()
	loop := true
	for loop {
		select {
		case event := <-events:
			switch event.EventType {
			case "click":
				if err := a.actionRepo.AddClickAction(event.PageURL, event.SlotID, event.BannerID, event.UserAge, event.UserSex); err != nil {
					a.logger.Log(ctx, errors.Wrapf(err, ErrProcessClickEvent, event.BannerID, event.PageURL, event.SlotID))
				}
			case "show":
				if err := a.actionRepo.AddShowAction(event.PageURL, event.SlotID, event.BannerID, event.UserAge, event.UserSex); err != nil {
					a.logger.Log(ctx, errors.Wrapf(err, ErrProcessShowEvent, event.BannerID, event.PageURL, event.SlotID))
				}
			}
		case <-ctx.Done():
			loop = false
			break
		}
	}
}

func (a *AggregatorInteractor) ListenEvents(ctx context.Context) error {
	events := make(chan entities.Event)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go a.processEvent(ctx, wg, events)
	if err := a.queue.Pull(ctx, events); err != nil {
		close(events)
		return errors.Wrapf(err, "could't fetch messages from queue")
	}
	wg.Wait()
	return nil
}
