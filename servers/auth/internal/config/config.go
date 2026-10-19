package config

import (
	"time"

	"auth/internal/config/internal/env"
)

type Database struct {
	User     string
	Name     string
	Password string
	Port     int
	Host     string
}
type JWT struct {
	PrivateKeyPath string
	TTL            time.Duration
}

type Config struct {
	Port           int
	DB             Database
	Cost           int
	UserServerAddr string
	JWT            JWT
}

func Load() (Config, error) {
	r := env.Reader{}
	c := Config{
		Port: r.IntRange("PORT", 1, 65535),

		DB: Database{
			User:     r.String("DB_USER"),
			Name:     r.String("DB_NAME"),
			Password: r.String("DB_PASSWORD"),
			Port:     r.IntRange("DB_PORT", 1, 65535),
			Host:     r.String("DB_HOST"),
		},
		Cost:           r.IntRange("HASH_COST", 4, 10),
		UserServerAddr: r.String("USER_SERVER_ADDR"),
		JWT: JWT{
			PrivateKeyPath: r.String("JWT_PRIVATE_KEY_PATH"),
			TTL:            r.DurationRange("JWT_TTL", time.Minute, 24*time.Hour),
		},
	}

	if err := r.Err(); err != nil {
		return Config{}, err
	}

	return c, nil
}
