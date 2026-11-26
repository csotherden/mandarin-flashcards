package auth

import "testing"

func TestValidateSecret(t *testing.T) {
	if err := ValidateSecret("JWT_SECRET", ""); err == nil {
		t.Fatal("expected empty secret to fail")
	}
	if err := ValidateSecret("JWT_SECRET", "changeme"); err == nil {
		t.Fatal("expected placeholder to fail")
	}
	if err := ValidateSecret("JWT_SECRET", "short-but-not-placeholder"); err == nil {
		t.Fatal("expected short secret to fail")
	}
	if err := ValidateSecret("JWT_SECRET", "this-is-a-long-enough-local-signing-secret"); err != nil {
		t.Fatal(err)
	}
}
