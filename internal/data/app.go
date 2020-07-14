package data

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shipa988/banner_rotator/internal/data/repository"
	"github.com/shipa988/banner_rotator/internal/domain/entities"
	"github.com/shipa988/banner_rotator/internal/domain/usecase"
	"os"
)

type App struct {
}

func NewApp() *App {
	return &App{}
}

func (a *App) Run(cfg *AppConfig, isDebug bool) (err error) {
	ctx:=context.Background()
	wr := os.Stdout
	if !isDebug {
		wr, err = os.OpenFile(cfg.Log.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return errors.Wrapf(err, "can't create/open log file")
		}
	}

	logger, err := NewLogger(wr)
	if err != nil {
		return errors.Wrapf(err, "can't init logger")
	}
	InitDB(ctx,cfg,logger)
	/*repo, err := InitRepo(cfg, logger)
	if err != nil {
		return errors.Wrapf(err, "can't init repository")
	}
	calendar := usecases.NewCalendar(repo, nil, logger)
	// set executors for api.
	apiHandler := httpservice.NewApiHandler(calendar, logger)
	// prepare http handler with all middlewares for server.
	httpHandler := httpserver.GetHandler(logger,apiHandler)

	wg := &sync.WaitGroup{}
	// prepare http server with handler
	httpServer := httpserver.NewHTTPServer(wg, logger, httpHandler)
	grpcServer := grpcserver.NewGRPCServer(wg, logger, calendar)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	wg.Add(1)
	go httpServer.Serve(net.JoinHostPort("localhost", cfg.API.HTTPPort))

	wg.Add(1)
	go grpcServer.Serve(net.JoinHostPort("localhost", cfg.API.GRPCPort))

	go func() {
		wg.Wait()
		quit<-os.Interrupt
	}()

	<-quit

	httpServer.StopServe()
	grpcServer.StopServe()*/

	return nil
}

/*
func InitRepo(cfg *AppConfig, logger usecase.Logger) (entities.EventRepo, error) {
	switch cfg.RepoType {
	case "db":
		repo, err := db.NewDBEventRepo(cfg.DB.Driver, cfg.DB.DSN, logger)
		if err != nil {
			return nil, errors.Wrapf(err, "can't init db repository")
		}
		return repo, nil
	case "inmemory":
		repo, err := inmemory.NewInMemoryEventRepo(inmemory.NewMapRepo(), logger)
		if err != nil {
			return nil, errors.Wrapf(err, "can't init inmemo repository")
		}
		return repo, nil
	default:
		return nil, errors.New("unknown repository type. I know next types:db-database,inmemory-map struct into app")
	}
}*/
func InitDB(ctx context.Context,cfg *AppConfig, logger usecase.Logger) {
	db, err := gorm.Open(cfg.DB.Dialect,  cfg.DB.DSN)
	if err != nil {
		fmt.Print(err)
	}
	db.Debug().AutoMigrate(&repository.Banner{}, &repository.Slot{}, &repository.Page{}, &repository.Group{}, &repository.BannerClick{}) //Миграция базы данных
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
	logger.Log(ctx,"db init")
}
