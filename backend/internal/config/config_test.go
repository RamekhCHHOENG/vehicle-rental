package config

import (
	"strings"
	"testing"
)

// The repository is public, so every placeholder in it is known to everyone.
// Load must refuse to start rather than run with a guessable access secret.
func TestLoadRejectsPublicAndWeakSecrets(t *testing.T) {
	strongJWT := "s6ZQ2rVx9pLmT4hK8wNfYbCe3JdUg7Aq"
	strongAdmin := "hZ4tQ9wLmXc2"

	tests := []struct {
		name        string
		jwt         string
		admin       string
		wantErrOn   string
		wantSuccess bool
	}{
		{"strong values load", strongJWT, strongAdmin, "", true},
		{"missing jwt secret", "", strongAdmin, "JWT_SECRET", false},
		{"missing admin password", strongJWT, "", "ADMIN_PASSWORD", false},
		{"jwt secret from .env.example", "change-me-to-a-long-random-string", strongAdmin, "JWT_SECRET", false},
		{"jwt secret from the old code default", "dev-secret-do-not-use-in-production", strongAdmin, "JWT_SECRET", false},
		{"admin password from .env.example", strongJWT, "admin123", "ADMIN_PASSWORD", false},
		{"jwt secret too short", "short", strongAdmin, "JWT_SECRET", false},
		{"admin password too short", strongJWT, "abc", "ADMIN_PASSWORD", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET", tt.jwt)
			t.Setenv("ADMIN_PASSWORD", tt.admin)

			cfg, err := Load()

			if tt.wantSuccess {
				if err != nil {
					t.Fatalf("expected config to load, got: %v", err)
				}
				if cfg.JWTSecret != tt.jwt {
					t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, tt.jwt)
				}
				return
			}

			if err == nil {
				t.Fatal("expected an error, config loaded instead")
			}
			if !strings.Contains(err.Error(), tt.wantErrOn) {
				t.Errorf("error should name %s, got: %v", tt.wantErrOn, err)
			}
		})
	}
}
