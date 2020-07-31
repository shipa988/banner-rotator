package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/shipa988/banner_rotator/cmd/rotator/internal/algorithms/multiarms"
	"github.com/shipa988/banner_rotator/cmd/rotator/internal/algorithms/random"
	"github.com/shipa988/banner_rotator/cmd/rotator/internal/data/controllers/grpcservice"
	"github.com/shipa988/banner_rotator/cmd/rotator/internal/domain/usecase"
	"github.com/shipa988/banner_rotator/internal/data/controllers/queueservice/kafkaservice"
	"github.com/shipa988/banner_rotator/internal/data/logger"
	"github.com/shipa988/banner_rotator/internal/data/logger/zaplogger"
	"github.com/shipa988/banner_rotator/internal/data/repository"
	"github.com/shipa988/banner_rotator/internal/domain/entities"
)

const ErrAppInit = "can't init app"
const ErrAppRun = "can't run app"
const ErrUpDB = "can't up db"
const ErrDownDB = "can't down db"

type App struct {
}

func NewApp() *App {
	return &App{}
}

func (a *App) initRotator(cfg *Config, isDebug, upDB bool, logger logger.Logger) (rotator usecase.Rotator, err error) {
	repo, err := initRepo(cfg, logger, isDebug)
	if err != nil {
		return nil, errors.Wrapf(err, ErrAppInit)
	}

	if upDB {
		if err := a.DBUp(cfg, isDebug); err != nil {
			return nil, errors.Wrapf(err, ErrAppInit)
		}
	}

	algo, err := initAlgo(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, ErrAppInit)
	}

	broker, err := initQueueBroker(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, ErrAppInit)
	}

	rotator, err = usecase.NewRotatorInteractor(repo, broker, algo, logger)
	if err != nil {
		return nil, errors.Wrapf(err, ErrAppInit)
	}
	err = rotator.Init()
	if err != nil {
		return nil, errors.Wrapf(err, ErrAppInit)
	}

	return
}

func (a *App) Run(cfg *Config, isDebug, upDB bool) (err error) {
	fmt.Println(cfg)
	logger, err := initLogger(cfg, isDebug)
	if err != nil {
		return errors.Wrapf(err, ErrAppRun)
	}

	rotator, err := a.initRotator(cfg, isDebug, upDB, logger)
	if err != nil {
		return errors.Wrapf(err, ErrAppRun)
	}

	wg := &sync.WaitGroup{}
	grpcServer := grpcservice.NewGRPCServer(wg, logger, rotator)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	wg.Add(1)
	go func() {
		l := grpcServer.PrepareListener(net.JoinHostPort("0.0.0.0", cfg.API.GRPCPort))
		if err := grpcServer.Serve(l); err != nil {
			logger.Log(context.Background(), errors.Wrap(err, ErrAppRun))
			quit <- os.Interrupt
		}
	}()

	wg.Add(1)
	go func() {
		if err := grpcServer.ServeGW(net.JoinHostPort("0.0.0.0", cfg.API.GRPCPort), net.JoinHostPort("0.0.0.0", cfg.API.GRPCGWPort)); err != nil {
			logger.Log(context.Background(), errors.Wrap(err, ErrAppRun))
			quit <- os.Interrupt
		}
	}()

	go func() {
		wg.Wait()
		quit <- os.Interrupt
	}()

	<-quit

	grpcServer.StopServe()
	grpcServer.StopGWServe()

	return nil
}

func (a *App) DBUp(cfg *Config, isDebug bool) error {
	logger, err := initLogger(cfg, isDebug)
	if err != nil {
		return errors.Wrapf(err, ErrUpDB)
	}

	repo, err := initRepo(cfg, logger, isDebug)
	if err != nil {
		return errors.Wrapf(err, ErrUpDB)
	}

	repo.CreateDB()
	return nil
}

func (a *App) DBDown(cfg *Config, isDebug bool) error {
	logger, err := initLogger(cfg, isDebug)
	if err != nil {
		return errors.Wrapf(err, ErrDownDB)
	}

	repo, err := initRepo(cfg, logger, isDebug)
	if err != nil {
		return errors.Wrapf(err, ErrDownDB)
	}

	repo.DeleteDB()
	return nil
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

func initQueueBroker(cfg *Config) (entities.EventQueue, error) {
	switch cfg.Queue.Name {
	case "kafka":
		broker := kafkaservice.NewKafkaManager(cfg.Kafka.Addr, cfg.Kafka.Topic)
		broker.InitWriter()
		return broker, nil
	case "rabbit":
		return nil, errors.New(`rabbit broker does't support'`)
	default:
		return nil, errors.New(`unknown queue broker`)
	}
}

func initAlgo(cfg *Config) (usecase.NextBannerAlgo, error) {
	switch cfg.Algo.Name {
	case "ucb1":
		algo := multiarms.NewUCB1Algo()
		return algo, nil
	case "thompson":
		algo := multiarms.NewUCB1Algo() //todo: realize this algo if necessary
		return algo, nil
	case "random":
		algo := random.NewRandomizer()
		return algo, nil
	default:
		return nil, errors.New(`unknown algorithm for returning banner id to show. I know Multi-armed_bandit realization algorithms: "ucb1"-https://en.wikipedia.org/wiki/Multi-armed_bandit,"random"-random id banner`)
	}
}

func initDB(cfg *Config) (*gorm.DB, error) {
	db, err := gorm.Open(cfg.DB.Dialect, cfg.DB.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "can't initialize db")
	}
	return db, nil
}
