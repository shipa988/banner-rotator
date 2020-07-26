package app

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shipa988/banner_rotator/cmd/aggregator/internal/domain/usecase"
	"github.com/shipa988/banner_rotator/internal/data/controllers/queueservice/kafkaservice"
	"github.com/shipa988/banner_rotator/internal/data/logger/zapLogger"
	"github.com/shipa988/banner_rotator/internal/data/repository"
	"github.com/shipa988/banner_rotator/internal/domain/entities"
	"net"
	"os"
	"sync"
)

type App struct {
}

func NewApp() *App {
	return &App{}
}
func (a *App) Run(cfg *AppConfig, isDebug bool) (err error) {
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wr := os.Stdout
	if !isDebug {
		wr, err = os.OpenFile(cfg.Log.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return errors.Wrapf(err, "can't create/open log file")
		}
	}
	logger, err := zapLogger.NewLogger(wr, isDebug)
	if err != nil {
		return errors.Wrapf(err, "can't init logger")
	}

	db, err := gorm.Open(cfg.DB.Dialect, cfg.DB.DSN)
	if err != nil {
		logger.Log(ctx, err.Error())
	}
	db.SetLogger(logger)
	repo := repository.NewPGRepo(db, isDebug)

	broker, err := InitQueueBroker(cfg)
	if err != nil {
		return errors.Wrapf(err, "can't queue manager")
	}
	aggregator, err := usecase.NewAggregatorInteractor(repo, broker)

	if err != nil {
		logger.Log(ctx, err.Error())
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := aggregator.ListenEvents(ctx); err != nil {
			logger.Log(ctx, err)
		}
	}()
	wg.Wait()
	return nil
}
func InitQueueBroker(cfg *AppConfig) (entities.EventQueue, error) {
	switch cfg.Queue.Name {
	case "kafka":
		broker := kafkaservice.NewKafkaManager(net.JoinHostPort("localhost", cfg.Kafka.Port), cfg.Kafka.Topic)
		broker.InitReader(cfg.Kafka.ConsumerGroup, cfg.Kafka.MinSize, cfg.Kafka.MaxSize)
		return broker, nil
	case "rabbit":
		return nil, errors.New(`rabbit broker does't support'`)
	default:
		return nil, errors.New(`unknown queue broker`)
	}
}
