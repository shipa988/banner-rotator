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
	"github.com/shipa988/banner_rotator/internal/data/logger/zapLogger"
	"github.com/shipa988/banner_rotator/internal/data/repository"
	"github.com/shipa988/banner_rotator/internal/domain/entities"
)

type App struct {
}

func NewApp() *App {
	return &App{}
}

func (a *App) Run(cfg *AppConfig, isDebug bool) (err error) {
	ctx := context.Background()
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
	db := InitDB(ctx, cfg, logger)
	db.SetLogger(logger)

	repo := repository.NewPGRepo(db, isDebug)

	algo, err := InitAlgo(cfg)
	if err != nil {
		return errors.Wrapf(err, "can't init core algorithm")
	}
	broker, err := InitQueueBroker(cfg)
	if err != nil {
		return errors.Wrapf(err, "can't queue manager")
	}
	rotator, err := usecase.NewRotatorInteractor(repo, broker, algo, logger)
	if err != nil {
		return errors.Wrapf(err, "can't create NewRotatorInteractor")
	}
	err = rotator.Init()
	if err != nil {
		return errors.Wrapf(err, "can't init NewRotatorInteractorr")
	}
	wg := &sync.WaitGroup{}

	grpcServer := grpcservice.NewGRPCServer(wg, logger, rotator)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	wg.Add(1)
	l := grpcServer.PrepareListener(net.JoinHostPort("localhost", cfg.API.GRPCPort))
	go grpcServer.Serve(l)

	wg.Add(1)
	go grpcServer.ServeGW(net.JoinHostPort("localhost", cfg.API.GRPCPort), net.JoinHostPort("localhost", cfg.API.GRPCGWPort))

	go func() {
		wg.Wait()
		quit <- os.Interrupt
	}()

	<-quit

	grpcServer.StopServe()
	grpcServer.StopGWServe()

	return nil
}
func InitQueueBroker(cfg *AppConfig) (entities.EventQueue, error) {
	switch cfg.Queue.Name {
	case "kafka":
		broker := kafkaservice.NewKafkaManager(net.JoinHostPort("localhost", cfg.Kafka.Port), cfg.Kafka.Topic)
		broker.InitWriter()
		return broker, nil
	case "rabbit":
		return nil, errors.New(`rabbit broker does't support'`)
	default:
		return nil, errors.New(`unknown queue broker`)
	}
}

func InitAlgo(cfg *AppConfig) (usecase.NextBannerAlgo, error) {
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

func InitDB(ctx context.Context, cfg *AppConfig, logger logger.Logger) *gorm.DB {
	db, err := gorm.Open(cfg.DB.Dialect, cfg.DB.DSN)
	if err != nil {
		fmt.Print(err)
	}
	db.Debug().AutoMigrate(&repository.Banner{}, &repository.Slot{}, &repository.Page{}, &repository.Group{}, &repository.BannerEvent{}, &repository.BannerSlot{}) //Миграция базы данных
	groups := []*entities.Group{
		{
			Description: "young man",
			Sex:         "man",
			MinAge:      0,
			MaxAge:      40,
		},
		{
			Description: "young women",
			Sex:         "women",
			MinAge:      0,
			MaxAge:      40,
		},
		{
			Description: "middle-age man",
			Sex:         "man",
			MinAge:      41,
			MaxAge:      60,
		},
		{
			Description: "middle-age women",
			Sex:         "women",
			MinAge:      41,
			MaxAge:      60,
		},
		{
			Description: "old man",
			Sex:         "man",
			MinAge:      61,
			MaxAge:      150,
		},
		{
			Description: "old women",
			Sex:         "women",
			MinAge:      61,
			MaxAge:      150,
		},
		{
			Description: "unknown age-sex group",
			Sex:         "unknown",
			MinAge:      0,
			MaxAge:      0,
		},
	}

	for _, group := range groups {
		db.Save(group)
	}
	logger.Log(ctx, "db init")
	return db
}
