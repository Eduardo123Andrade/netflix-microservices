package config

import "auth/internal/config/internal/env"

type Config struct {
	Port int
}

func Load() (Config, error) {
	r := env.Reader{}
	c := Config{
		Port: r.IntRange("PORT", 1, 65535),
	}

	if err := r.Err(); err != nil {
		return Config{}, err
	}

	return c, nil
}
