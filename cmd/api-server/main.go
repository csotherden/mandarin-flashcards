package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"entgo.io/ent/dialect/sql/schema"
	"github.com/csotherden/mandarin-flashcards/ent"
	"github.com/csotherden/mandarin-flashcards/internal/auth"
	"github.com/csotherden/mandarin-flashcards/internal/database"
	"github.com/csotherden/mandarin-flashcards/internal/library"
	"github.com/csotherden/mandarin-flashcards/internal/logging"
	"github.com/csotherden/mandarin-flashcards/internal/server"
	"github.com/csotherden/mandarin-flashcards/internal/unsplash"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const gracefulShutdownTimeout = 15 * time.Second

var (
	srv    *server.Server
	logger *zap.Logger
	client *ent.Client
	sqlDB  *sql.DB
)

func envBool(name string) (bool, bool) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return false, false
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, false
	}
	return parsed, true
}

func init() {
	var envFile, libraryFile string
	flag.StringVar(&envFile, "env", ".env", "env file path")
	flag.StringVar(&libraryFile, "library-load", "", "load global library sets from this JSON file (- for stdin), then exit")
	flag.Parse()

	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: Error loading env file: %v", err)
	}

	var err error
	logger, err = logging.NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	config := server.NewConfig()
	if err := auth.ValidateSecret("JWT_SECRET", os.Getenv("JWT_SECRET")); err != nil {
		logger.Fatal("Invalid JWT_SECRET", zap.Error(err), zap.Int("minimum_length", auth.MinSecretLen))
	}
	if err := auth.ValidateSecret("JWT_REFRESH_SECRET", os.Getenv("JWT_REFRESH_SECRET")); err != nil {
		logger.Fatal("Invalid JWT_REFRESH_SECRET", zap.Error(err), zap.Int("minimum_length", auth.MinSecretLen))
	}

	if destructive, ok := envBool("DESTRUCTIVE_MIGRATION"); ok && destructive && !config.IsExplicitLocalEnv() {
		logger.Fatal("DESTRUCTIVE_MIGRATION=true requires ENV explicitly set to development, test, or local",
			zap.String("environment", config.Environment),
			zap.Bool("environment_explicit", config.EnvironmentExplicit),
		)
	}

	client, sqlDB, err = database.NewClient(logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	autoMigrate, autoMigrateSet := envBool("AUTO_MIGRATE")
	destructiveMigration, destructiveMigrationSet := envBool("DESTRUCTIVE_MIGRATION")
	if !autoMigrateSet && os.Getenv("AUTO_MIGRATE") != "" {
		logger.Warn("AUTO_MIGRATE has invalid value, defaulting to false", zap.String("value", os.Getenv("AUTO_MIGRATE")))
	}
	if !destructiveMigrationSet && os.Getenv("DESTRUCTIVE_MIGRATION") != "" {
		logger.Warn("DESTRUCTIVE_MIGRATION has invalid value, defaulting to false", zap.String("value", os.Getenv("DESTRUCTIVE_MIGRATION")))
	}

	if autoMigrate {
		if err := client.Schema.Create(context.Background()); err != nil {
			logger.Fatal("Failed to run database migrations", zap.Error(err))
		}
		logger.Debug("Database migrations applied")
		if destructiveMigration {
			logger.Warn("Running destructive database migration options because DESTRUCTIVE_MIGRATION=true")
			if err := client.Schema.Create(context.Background(), schema.WithDropColumn(true), schema.WithDropIndex(true)); err != nil {
				logger.Fatal("Failed to run destructive database migrations", zap.Error(err))
			}
		}
	}

	if libraryFile != "" {
		os.Exit(loadLibrary(libraryFile))
	}

	srv, err = server.NewServer(config, logger, client, sqlDB)
	if err != nil {
		logger.Fatal("Failed to create server", zap.Error(err))
	}
}

func main() {
	defer logger.Sync()
	defer client.Close()

	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	sig := <-signals
	logger.Info("Received signal, shutting down server", zap.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
	logger.Info("Server gracefully stopped")
}

// loadLibrary runs `api-server -library-load file.json` (or - for stdin): it makes the global
// library match the file, set by set, and exits. Safe to rerun; sets are matched by slug.
func loadLibrary(path string) int {
	in := os.Stdin
	if path != "-" {
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		defer f.Close()
		in = f
	}
	res, err := library.Load(context.Background(), client, in)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("library loaded: sets %d new, %d updated; cards %d new, %d updated, %d removed\n",
		res.SetsCreated, res.SetsUpdated, res.CardsCreated, res.CardsUpdated, res.CardsRemoved)

	// Unsplash asks to be told when a photo is used; for the library that's when it's loaded.
	photos := unsplash.New(strings.TrimSpace(os.Getenv("UNSPLASH_ACCESS_KEY")), strings.TrimSpace(os.Getenv("UNSPLASH_API_URL")))
	if len(res.NewPhotos) > 0 && !photos.Enabled() {
		fmt.Printf("unsplash: not configured, %d new photos not reported\n", len(res.NewPhotos))
		return 0
	}
	failed := 0
	for _, location := range res.NewPhotos {
		if err := photos.TrackDownload(context.Background(), location); err != nil {
			failed++
			fmt.Fprintf(os.Stderr, "unsplash: %s: %v\n", location, err)
		}
	}
	if len(res.NewPhotos) > 0 {
		fmt.Printf("unsplash: reported %d new photos (%d failed)\n", len(res.NewPhotos)-failed, failed)
	}
	return 0
}
