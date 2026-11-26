package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// FileStore is the local object store. This deployment does not use S3.
type FileStore interface {
	UploadFile(ctx context.Context, key string, content []byte) error
	DownloadFile(ctx context.Context, key string) ([]byte, error)
	DeleteFile(ctx context.Context, key string) error
}

type localFileStore struct {
	dataDir string
	logger  *zap.Logger
}

// NewLocalFileStore persists objects under dataDir.
func NewLocalFileStore(dataDir string, logger *zap.Logger) (FileStore, error) {
	if strings.TrimSpace(dataDir) == "" {
		return nil, fmt.Errorf("file store directory is required")
	}
	if err := os.MkdirAll(dataDir, 0o775); err != nil {
		return nil, fmt.Errorf("create file store directory: %w", err)
	}
	logger.Info("Using local file store", zap.String("dir", dataDir))
	return &localFileStore{dataDir: dataDir, logger: logger}, nil
}

func (l *localFileStore) path(key string) (string, error) {
	clean := filepath.Clean(key)
	if clean == "." || strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return "", fmt.Errorf("invalid object key")
	}
	for _, part := range strings.Split(clean, string(filepath.Separator)) {
		if part == ".." {
			return "", fmt.Errorf("invalid object key")
		}
	}
	full := filepath.Join(l.dataDir, clean)
	rel, err := filepath.Rel(l.dataDir, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid object key")
	}
	return full, nil
}

func (l *localFileStore) UploadFile(_ context.Context, key string, content []byte) error {
	full, err := l.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o775); err != nil {
		return fmt.Errorf("local file store mkdir for %s: %w", key, err)
	}
	tmp := full + ".tmp"
	if err := os.WriteFile(tmp, content, 0o640); err != nil {
		return fmt.Errorf("local file store write %s: %w", key, err)
	}
	if err := os.Rename(tmp, full); err != nil {
		return fmt.Errorf("local file store rename %s: %w", key, err)
	}
	return nil
}

func (l *localFileStore) DownloadFile(_ context.Context, key string) ([]byte, error) {
	full, err := l.path(key)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(full)
	if err != nil {
		return nil, fmt.Errorf("local file store read %s: %w", key, err)
	}
	return content, nil
}

func (l *localFileStore) DeleteFile(_ context.Context, key string) error {
	full, err := l.path(key)
	if err != nil {
		return err
	}
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("local file store delete %s: %w", key, err)
	}
	return nil
}
