package usecase

import (
	"context"
)

type Aggregator interface {
	ListenEvents(ctx context.Context) error
}
