package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/csotherden/mandarin-flashcards/ent"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// NewClient opens Postgres and returns the Ent client plus the underlying pool
// so health checks can ping it. Ent does not expose Ping.
func NewClient(logger *zap.Logger) (*ent.Client, *sql.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		return nil, nil, fmt.Errorf("DB_PASSWORD environment variable is not set")
	}

	dbType := strings.TrimSpace(os.Getenv("DB_TYPE"))
	if dbType == "" {
		dbType = "postgres"
	}
	if dbType != "postgres" {
		return nil, nil, fmt.Errorf("invalid database type: %q", dbType)
	}

	// Fail closed: an unset DB_SSL_MODE must not silently mean no TLS.
	// Local docker-compose Postgres has no TLS listener, so it sets
	// DB_SSL_MODE=disable explicitly.
	sslMode := os.Getenv("DB_SSL_MODE")
	if sslMode == "" {
		sslMode = "require"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPass, dbName, sslMode)

	drv, err := entsql.Open(dialect.Postgres, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed opening connection to postgres: %w", err)
	}

	db := drv.DB()
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(50)
	db.SetConnMaxLifetime(time.Hour)

	client := ent.NewClient(ent.Driver(drv))
	logger.Info("Database connection established", zap.String("host", dbHost), zap.String("database", dbName))
	return client, db, nil
}
