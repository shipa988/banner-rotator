package app

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/shipa988/banner_rotator/cmd/aggregator/internal/domain/usecase"
	"github.com/shipa988/banner_rotator/internal/data/controllers/queueservice/kafkaservice"
	"github.com/shipa988/banner_rotator/internal/data/logger"
	"github.com/shipa988/banner_rotator/internal/data/logger/zaplogger"
	"github.com/shipa988/banner_rotator/internal/data/repository"
	"github.com/shipa988/banner_rotator/internal/domain/entities"
)

const ErrAppRun = "can't run app"

type App struct {
}

func NewApp() *App {
	return &App{}
}
func (a *App) Run(cfg *Config, isDebug bool) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt)

	logger, err := initLogger(cfg, isDebug)
	if err != nil {
		return errors.Wrapf(err, ErrAppRun)
	}

	repo, err := initRepo(cfg, logger, isDebug)
	if err != nil {
		return errors.Wrapf(err, ErrAppRun)
	}

	broker, err := initQueueBroker(cfg)
	if err != nil {
		return errors.Wrapf(err, "can't queue manager")
	}
	aggregator, err := usecase.NewAggregatorInteractor(repo, broker, logger)

	if err != nil {
		logger.Log(ctx, err.Error())
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Log(ctx, "listen events from queue...")
		if err := aggregator.ListenEvents(ctx); err != nil {
			logger.Log(ctx, err)
		}
	}()

	go func() {
		wg.Wait()
		done <- os.Interrupt
	}()

	<-done
	return nil
}
func initQueueBroker(cfg *Config) (entities.EventQueue, error) {
	switch cfg.Queue.Name {
	case "kafka":
		broker := kafkaservice.NewKafkaManager(cfg.Kafka.Addr, cfg.Kafka.Topic)
		broker.InitReader(cfg.Kafka.ConsumerGroup, cfg.Kafka.MinSize, cfg.Kafka.MaxSize)
		return broker, nil
	case "rabbit":
		return nil, errors.New(`rabbit broker does't support'`)
	default:
		return nil, errors.New(`unknown queue broker`)
	}
}

func initRepo(cfg *Config, logger logger.Logger, isDebug bool) (repo *repository.PGRepo, err error) {
	db, err := initDB(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "can't init repository")
	}
	repo = repository.NewPGRepo(db, logger, isDebug)
	return repo, nil
}

func initLogger(cfg *Config, isDebug bool) (logger logger.Logger, err error) {
	wr := os.Stdout
	if !isDebug {
		wr, err = os.OpenFile(cfg.Log.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, errors.Wrapf(err, "can't create/open log file")
		}
	}
	logger = zaplogger.NewLogger(wr, isDebug)
	return logger, nil
}

func initDB(cfg *Config) (*gorm.DB, error) {
	db, err := gorm.Open(cfg.DB.Dialect, cfg.DB.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "can't initialize db")
	}
	return db, nil
}
