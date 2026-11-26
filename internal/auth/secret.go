package auth

import (
	"errors"
	"fmt"
	"strings"
)

// MinSecretLen is the minimum length for JWT signing secrets.
const MinSecretLen = 32

var knownPlaceholders = []string{
	"devsecret",
	"devrefresh",
	"changeme",
	"changeit",
	"secret",
	"password",
	"example",
	"test",
	"jwtsecret",
	"your_super_secret_key_change_this_in_production",
	"change_this_secret_to_a_long_random_value",
}

var (
	errSecretNotConfigured = errors.New("secret is not configured")
	errSecretTooShort      = fmt.Errorf("secret must be at least %d characters", MinSecretLen)
)

// ValidateSecret rejects empty, short, and known-placeholder signing secrets.
func ValidateSecret(name, value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("%s: %w", name, errSecretNotConfigured)
	}
	lower := strings.ToLower(trimmed)
	for _, placeholder := range knownPlaceholders {
		if lower == placeholder {
			return fmt.Errorf("%s: refusing known placeholder value %q", name, placeholder)
		}
	}
	if len(trimmed) < MinSecretLen {
		return fmt.Errorf("%s: %w", name, errSecretTooShort)
	}
	return nil
}
