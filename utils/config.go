package utils

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret      string
	DatabaseDSN    string
	ClientUsername string
	ClientPassword string
}
func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		JWTSecret:      strings.TrimSpace(os.Getenv("JWT_SECRET_KEY")),
		DatabaseDSN:    strings.TrimSpace(os.Getenv("DATABASE_SECRET_KEY")),
		ClientUsername: strings.TrimSpace(os.Getenv("CLIENT_USERNAME")),
		ClientPassword: strings.TrimSpace(os.Getenv("CLIENT_PASSWORD")),
	}
	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}

	if cfg.DatabaseDSN == "" {
		log.Printf("DATABASE_SECRET_KEY is not set")
	}
	if cfg.JWTSecret == "" {
		log.Printf("JWT_SECRET_KEY is not set")
	}
	return cfg, nil
}
