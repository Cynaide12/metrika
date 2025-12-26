package main

import (
	"log/slog"
	"metrika/internal/config"
	"metrika/internal/domain/analytics"
	"metrika/internal/domain/auth"
	"metrika/internal/infrastructure/jwt"
	"metrika/internal/infrastructure/logger"
	"metrika/internal/infrastructure/mock"
	"metrika/internal/infrastructure/postgres"
	sessionworker "metrika/internal/infrastructure/session_worker"
	"metrika/internal/infrastructure/tracker"
	analhandler "metrika/internal/transport/http/v1/analytics"
	authhandler "metrika/internal/transport/http/v1/auth"
	methandler "metrika/internal/transport/http/v1/metrika"
	mid "metrika/internal/transport/http/v1/middleware"
	analuc "metrika/internal/usecase/analytics"
	authuc "metrika/internal/usecase/auth"
	"metrika/internal/usecase/metrika"
	"metrika/pkg/logger/sl"

	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/robfig/cron/v3"
)

type repos struct {
	domains        analytics.DomainRepository
	events         analytics.EventsRepository
	guests         analytics.GuestsRepository
	guest_sessions analytics.GuestSessionRepository
	sessions       auth.SessionRepository
	users          auth.UserRepository
}

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
	guest_sessions := postgres.NewGuestSessionRepository(db)
	domains := postgres.NewDomainRepository(db)
	sessions := postgres.NewSessionRepository(db)
	guests := postgres.NewGuestsRepository(db)
	users := postgres.NewAuthRepository(db)
	tx := postgres.NewTxManager(db)

	repos := repos{
		domains:        domains,
		events:         events,
		sessions:       sessions,
		guests:         guests,
		guest_sessions: guest_sessions,
		users:          users,
	}

	tracker := tracker.New(1000, time.Second*15, 10000, events)

	cleanup_stale_sessions_uc := analuc.NewCleanupBatchSessionsUseCase(log, guest_sessions, tx)

	sessions_worker := sessionworker.NewSessionsWorker(log, time.Second*15, cleanup_stale_sessions_uc, make(chan struct{}))

	go sessions_worker.StartSessionManager()

	// setupMockGenerator(log, tracker, cfg, repos)

	log.Info("db connect succesful")

	log.Info("scheduler start succesful")

	setupRouter(cfg, log, tracker, tx, repos)
}

func setupMockGenerator(log *slog.Logger, tracker *tracker.Tracker, cfg *config.Config, repos repos) {
	mockGenerator := mock.NewGenerator()
	adapter := mock.MockServiceAdapter{Events: repos.events,
		Domains:  repos.domains,
		Sessions: repos.guest_sessions,
		Guests:   repos.guests}
	mockService := mock.NewMockService(adapter, mockGenerator, log, tracker, cfg.MockConfig)

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

func setupRouter(cfg *config.Config, log *slog.Logger, tracker *tracker.Tracker, tx *postgres.TxManager, repos repos) {
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
		AllowedOrigins: []string{"http://*, https://", "http://localhost:3000"},
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

	evuc := analuc.NewCollectEventsUseCase(repos.events, tracker, repos.guest_sessions, tx)
	createsesuc := analuc.NewGetGuestSessionUseCase(repos.guests, repos.guest_sessions, repos.domains, log)

	tokens := jwt.NewJwtProvider(cfg.JWTSecret)

	loginuc := authuc.NewLoginUseCase(repos.users, repos.sessions, tokens)
	refreshuc := authuc.NewRefreshUseCase(repos.sessions, tokens)
	registeruc := authuc.NewRegisterUseCase(repos.users, repos.sessions, tokens, log)
	guestSessionsByRangeDateuc := metrika.NewSessionsByRangeDateUseCase(repos.guest_sessions)
	activeSessionsuc := metrika.NewAciveSessionsUseCase(log, repos.guest_sessions)

	analyticsHandler := analhandler.NewHandler(log, evuc, createsesuc)
	authorizationHandler := authhandler.NewHandler(log, loginuc, refreshuc, registeruc)
	metrikaHandler := methandler.NewHandler(log, guestSessionsByRangeDateuc, activeSessionsuc)

	jwtProvider := jwt.NewJwtProvider(cfg.JWTSecret)

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(mid.AuthMiddleware(log, cfg.JWTSecret, *cfg, *jwtProvider))
			r.Route("/analytics", func(r chi.Router) {
				r.Post("/events", analyticsHandler.AddEvent)
				r.Post("/sessions", analyticsHandler.CreateGuestSession)
			})
			r.Route("/metrika", func(r chi.Router) {
				r.Route("/{domain_id}", func(r chi.Router) {
					r.Get("/guests/visits", metrikaHandler.GetGuestSessionByRangeDate)
					r.Get("/guests/online", metrikaHandler.GetCountActiveSessions)
				})
			})
		})
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authorizationHandler.Login)
			r.Put("/refresh", authorizationHandler.Refresh)
			r.Put("/logout", authorizationHandler.Logout)
			r.Post("/register", authorizationHandler.Register)
		})

	})

	log.Info("starting server", slog.String("address", srv.Addr))

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server", sl.Err(err))

		os.Exit(1)
	}

	log.Error("server stopped")

}
