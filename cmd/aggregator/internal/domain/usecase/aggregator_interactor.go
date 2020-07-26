package usecase

import (
	"context"
	"github.com/pkg/errors"
	"github.com/shipa988/banner_rotator/internal/domain/entities"
	"sync"
)

const (
	ErrProcessClickEvent = `can't register click event for banner id: "%v", page: "%v", slot id: "%v"`
	ErrProcessShowEvent  = `can't register show event for banner id: "%v", page: "%v", slot id: "%v"`
)

var _ Aggregator = (*AggregatorInteractor)(nil)

type AggregatorInteractor struct {
	actionRepo entities.ActionRepository
	queue      entities.EventQueue
}

func NewAggregatorInteractor(repo entities.ActionRepository, queueBroker entities.EventQueue) (*AggregatorInteractor, error) {
	return &AggregatorInteractor{
		actionRepo: repo,
		queue:      queueBroker,
	}, nil
}

func (a *AggregatorInteractor) processEvent(ctx context.Context, wg *sync.WaitGroup, events <-chan entities.Event) error {
	defer wg.Done()
	for {
		select {
		case event := <-events:
			switch t := event.EventType; {
			case t == "click":
				if err := a.actionRepo.AddAction(t, event.PageURL, event.SlotID, event.BannerID, event.UserAge, event.UserSex); err != nil {
					return errors.Wrapf(err, ErrProcessClickEvent, event.BannerID, event.PageURL, event.SlotID)
				}
			case t == "show":
				if err := a.actionRepo.AddAction(t, event.PageURL, event.SlotID, event.BannerID, event.UserAge, event.UserSex); err != nil {
					return errors.Wrapf(err, ErrProcessShowEvent, event.BannerID, event.PageURL, event.SlotID)
				}
			}
		case <-ctx.Done():
			//todo:close condext
		}
	}

	return nil
}

func (a *AggregatorInteractor) ListenEvents(ctx context.Context) error {
	events := make(chan entities.Event)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go a.processEvent(ctx, wg, events)
	if err := a.queue.Pull(events); err != nil { //todo:addcontext inside
		return errors.Wrapf(err, "could't fetch messages from queue")
	}
	wg.Wait()
	return nil
}
