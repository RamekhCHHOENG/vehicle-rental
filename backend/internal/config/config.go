package config

import "os"

type Config struct {
	Port          string
	DBURL         string
	JWTSecret     string
	AdminEmail    string
	AdminPassword string
}

func Load() Config {
	return Config{
		Port:          getenv("API_PORT", "8090"),
		DBURL:         getenv("DB_URL", "host=localhost port=5434 user=carrental password=carrental_dev_password dbname=carrental sslmode=disable"),
		JWTSecret:     getenv("JWT_SECRET", "dev-secret-do-not-use-in-production"),
		AdminEmail:    getenv("ADMIN_EMAIL", "admin@carrental.local"),
		AdminPassword: getenv("ADMIN_PASSWORD", "admin123"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
