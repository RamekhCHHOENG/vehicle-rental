package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port          string
	DBURL         string
	JWTSecret     string
	AdminEmail    string
	AdminPassword string
	// Origin the browser loads the site from, allowed through CORS. Behind a
	// reverse proxy the API is same-origin and this is unused.
	WebOrigin string
}

// publicValues are placeholders that appear in the public repository. A running
// service must never use one: the JWT secret would let anyone mint an admin
// token, and the admin password is the admin account itself. Rejecting them
// everywhere — not just in production — means there is no environment flag to
// get wrong.
var publicValues = map[string]bool{
	"change-me-to-a-long-random-string":   true,
	"dev-secret-do-not-use-in-production": true,
	"admin123":                            true,
	"carrental_dev_password":              true,
}

// Load reads configuration from the environment. It fails rather than falling
// back to a guessable default for anything that grants access.
func Load() (Config, error) {
	cfg := Config{
		Port:          getenv("API_PORT", "8090"),
		DBURL:         getenv("DB_URL", "host=localhost port=5434 user=carrental password=carrental_dev_password dbname=carrental sslmode=disable"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		AdminEmail:    getenv("ADMIN_EMAIL", "admin@carrental.local"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		WebOrigin:     getenv("WEB_ORIGIN", "http://localhost:3000"),
	}

	if err := requireSecret("JWT_SECRET", cfg.JWTSecret, 24); err != nil {
		return Config{}, err
	}
	if err := requireSecret("ADMIN_PASSWORD", cfg.AdminPassword, 8); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func requireSecret(name, value string, minLen int) error {
	switch {
	case value == "":
		return fmt.Errorf("%s is not set.\nGenerate one with:  openssl rand -base64 32", name)
	case publicValues[value]:
		return fmt.Errorf("%s is set to a placeholder from the public repository, so it is not secret.\nGenerate one with:  openssl rand -base64 32", name)
	case len(value) < minLen:
		return fmt.Errorf("%s must be at least %d characters (got %d)", name, minLen, len(value))
	}
	return nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
