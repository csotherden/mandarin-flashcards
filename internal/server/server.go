package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/internal/llm"
	"github.com/csotherden/mandarin-flashcards/internal/ratelimit"
	"github.com/csotherden/mandarin-flashcards/internal/storage"
	"github.com/csotherden/mandarin-flashcards/internal/toneperfect"
	"github.com/csotherden/mandarin-flashcards/internal/unsplash"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Server struct {
	config     *Config
	logger     *zap.Logger
	router     *mux.Router
	server     *http.Server
	db         *ent.Client
	sqlDB      *sql.DB
	files      storage.FileStore
	tones      *toneperfect.Client
	coach      *llm.Client
	photos     *unsplash.Client
	llmLimit   *ratelimit.Limiter
	imgLimit   *ratelimit.Limiter
	gradeLimit *ratelimit.Limiter
}

func NewServer(config *Config, logger *zap.Logger, db *ent.Client, sqlDB *sql.DB) (*Server, error) {
	files, err := storage.NewLocalFileStore(config.FileStoreDir, logger)
	if err != nil {
		return nil, err
	}
	s := &Server{
		config:     config,
		logger:     logger,
		router:     mux.NewRouter(),
		db:         db,
		sqlDB:      sqlDB,
		files:      files,
		tones:      toneperfect.New(config.TonePerfectAPIKey, config.TonePerfectAPIURL),
		coach:      llm.New(config.LLMAPIKey, config.LLMBaseURL, config.LLMModel),
		photos:     unsplash.New(config.UnsplashAccessKey),
		llmLimit:   ratelimit.New(20, time.Minute),
		imgLimit:   ratelimit.New(40, time.Minute),
		gradeLimit: ratelimit.New(30, time.Minute),
	}
	s.router.Use(s.corsMiddleware)
	s.router.PathPrefix("/").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	s.setupRoutes()
	s.router.NotFoundHandler = s.corsMiddleware(http.NotFoundHandler())
	s.router.MethodNotAllowedHandler = s.corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", config.Host, config.Port),
		Handler:      s.router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}
	return s, nil
}

func (s *Server) Start() error {
	s.logger.Info("Starting server", zap.String("address", s.server.Addr))
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) isAllowedOrigin(origin string) bool {
	for _, allowed := range s.config.AllowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && s.isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
