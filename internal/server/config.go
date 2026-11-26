package server

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Host                string
	Port                string
	Environment         string
	EnvironmentExplicit bool
	AllowedOrigins      []string
	FileStoreDir        string
	TonePerfectAPIKey   string
	TonePerfectAPIURL   string
	UnsplashAccessKey   string
	LLMAPIKey           string
	LLMBaseURL          string
	LLMModel            string
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	IdleTimeout         time.Duration
}

func NewConfig() *Config {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	envValue, envSet := os.LookupEnv("ENV")
	envValue = strings.ToLower(strings.TrimSpace(envValue))
	envSet = envSet && envValue != ""
	environment := envValue
	if environment == "" {
		environment = "development"
	}

	allowedOrigins := []string{
		"http://localhost:5173",
		"http://localhost:8091",
	}
	if raw := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS")); raw != "" {
		allowedOrigins = nil
		for _, origin := range strings.Split(raw, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowedOrigins = append(allowedOrigins, origin)
			}
		}
	}

	fileDir := strings.TrimSpace(os.Getenv("FILE_STORE_DIR"))
	if fileDir == "" {
		fileDir = "/var/lib/mandarin-files"
	}

	return &Config{
		Host:                host,
		Port:                port,
		Environment:         environment,
		EnvironmentExplicit: envSet,
		AllowedOrigins:      allowedOrigins,
		FileStoreDir:        fileDir,
		TonePerfectAPIKey:   strings.TrimSpace(os.Getenv("TONEPERFECT_API_KEY")),
		TonePerfectAPIURL:   strings.TrimSpace(os.Getenv("TONEPERFECT_API_URL")),
		UnsplashAccessKey:   strings.TrimSpace(os.Getenv("UNSPLASH_ACCESS_KEY")),
		LLMAPIKey:           strings.TrimSpace(os.Getenv("LLM_API_KEY")),
		LLMBaseURL:          strings.TrimSpace(os.Getenv("LLM_BASE_URL")),
		LLMModel:            strings.TrimSpace(os.Getenv("LLM_MODEL")),
		ReadTimeout:         60 * time.Second,
		WriteTimeout:        120 * time.Second,
		IdleTimeout:         120 * time.Second,
	}
}

// IsExplicitLocalEnv is true only when ENV was set to a local or test value.
// Destructive migrations refuse to run otherwise.
func (c *Config) IsExplicitLocalEnv() bool {
	if c == nil || !c.EnvironmentExplicit {
		return false
	}
	switch c.Environment {
	case "development", "test", "local":
		return true
	default:
		return false
	}
}
