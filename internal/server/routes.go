package server

import (
	"github.com/csotherden/mandarin-flashcards/internal/handlers/auth"
	"github.com/csotherden/mandarin-flashcards/internal/handlers/autofill"
	"github.com/csotherden/mandarin-flashcards/internal/handlers/cardset"
	"github.com/csotherden/mandarin-flashcards/internal/handlers/category"
	"github.com/csotherden/mandarin-flashcards/internal/handlers/flashcard"
	"github.com/csotherden/mandarin-flashcards/internal/handlers/generate"
	"github.com/csotherden/mandarin-flashcards/internal/handlers/health"
	"github.com/csotherden/mandarin-flashcards/internal/handlers/home"
	"github.com/csotherden/mandarin-flashcards/internal/handlers/images"
	"github.com/csotherden/mandarin-flashcards/internal/handlers/lesson"
	"github.com/csotherden/mandarin-flashcards/internal/handlers/libraryhandler"
	"github.com/csotherden/mandarin-flashcards/internal/handlers/study"
	userhandler "github.com/csotherden/mandarin-flashcards/internal/handlers/user"
	"github.com/csotherden/mandarin-flashcards/internal/middleware"
)

func (s *Server) setupRoutes() {
	api := s.router.PathPrefix("/api").Subrouter()

	health.NewHandler(s.sqlDB, s.logger).Register(api)
	auth.NewHandler(s.db, s.logger).Register(api)

	protected := api.NewRoute().Subrouter()
	protected.Use(middleware.AuthMiddleware(s.db, s.logger))

	userhandler.NewHandler(s.db, s.logger).Register(protected)
	home.NewHandler(s.db, s.logger).Register(protected)
	category.NewHandler(s.db, s.logger).Register(protected)
	autofill.NewHandler(s.db, s.coach, s.llmLimit, s.logger).Register(protected)
	cards := flashcard.NewHandler(s.db, s.files, s.photos, s.logger)
	cards.Register(protected)
	generate.NewHandler(s.db, s.coach, s.photos, cards, s.llmLimit, s.logger).Register(protected)
	images.NewHandler(s.photos, s.imgLimit, s.logger).Register(protected)
	cardset.NewHandler(s.db, s.logger).Register(protected)
	libraryhandler.NewHandler(s.db, s.logger).Register(protected)
	lesson.NewHandler(s.db, s.logger).Register(protected)
	study.NewHandler(s.db, s.tones, s.coach, s.gradeLimit, s.logger).Register(protected)
}
