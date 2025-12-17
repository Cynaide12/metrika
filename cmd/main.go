package main

import (
	"log/slog"
	"metrika/internal/config"
	"metrika/internal/infrastructure/logger"
	"metrika/internal/infrastructure/postgres"
	sessionworker "metrika/internal/infrastructure/session_worker"
	"metrika/internal/infrastructure/tracker"
	"metrika/pkg/logger/sl"

	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/robfig/cron/v3"
)


func main() {
	cfg := config.MustLoad()

	log, _, rotate, err := logger.SetupLogger(cfg.Env, cfg.LogFilePath)
	if err != nil {
		panic(err)
	}

	log.Info("starting metrika_server", slog.String("env", cfg.Env))

	log.Debug("debug messages are enabled")

	setupLogRotation(rotate)

	log.Info("logs rotation are enabled")

	db, err := postgres.New(cfg)
	if err != nil {
		log.Error("failed connect to db", sl.Err(err))
		os.Exit(1)
	}

	events := postgres.NewEventsRepository(db)

	tracker := tracker.New(1000, time.Second * 15, 10000, events)

	sessions_worker := sessionworker.NewSessionsWorker(log, time.Second*15, events)

	go sessions_worker.StartSessionManager()

	// setupMockGenerator(storage, log, tracker, cfg)

	log.Info("db connect succesful")

	log.Info("scheduler start succesful")

	initRouter(cfg, log, storage, tracker)
}

func setupMockGenerator(storage *repository.Repository, log *slog.Logger, tracker *tracker.Tracker, cfg *config.Config) {
	mockGenerator := mock.NewGenerator()

	mockService := service.NewMockService(storage, mockGenerator, log, tracker, cfg.MockConfig)

	go mockService.StartEventsGenerator()
}

func setupLogRotation(rotate func()) {
	//запускаем ротацию логов каждые сутки
	c := cron.New(cron.WithLocation(time.Local))

	c.AddFunc("@every 1d", func() {
		rotate()
	})

	c.Start()
}

func initRouter(cfg *config.Config, log *slog.Logger, storage *repository.Repository, tracker *tracker.Tracker) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(logger.New(log, cfg))
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      r,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://*", "https://*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            true,
	}))

	ms := service.New(storage, log, tracker)

	handlers.New(log, ms, r)

	log.Info("starting server", slog.String("address", srv.Addr))

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server", sl.Err(err))

		os.Exit(1)
	}

	log.Error("server stopped")

}
