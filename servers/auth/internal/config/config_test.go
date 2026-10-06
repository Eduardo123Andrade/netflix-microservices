package config

import (
	"auth/internal/config/internal/env"
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

var expect = Config{
	Port: 8080,
	DB: Database{
		User:     "test_user",
		Password: "test_password",
		Name:     "test_db",
		Port:     5450,
		Host:     "127.0.0.1",
	},
}

func initEnv(t *testing.T) {
	t.Helper()
	t.Setenv("PORT", strconv.Itoa(expect.Port))
	t.Setenv("DB_USER", expect.DB.User)
	t.Setenv("DB_PASSWORD", expect.DB.Password)
	t.Setenv("DB_NAME", expect.DB.Name)
	t.Setenv("DB_PORT", strconv.Itoa(expect.DB.Port))
	t.Setenv("DB_HOST", expect.DB.Host)
}

func TestLoad(t *testing.T) {
	initEnv(t)
	config, err := Load()

	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if !reflect.DeepEqual(config, expect) {
		t.Errorf("Load() = %+v; want %+v", config, expect)
	}
}

func TestLoadInvalidEnv(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		envValue string
		unset    bool
		wantErr  error
	}{
		{name: "PORT not integer", envName: "PORT", envValue: "ABC", wantErr: env.ErrNotInteger},
		{name: "PORT out of range", envName: "PORT", envValue: "0", wantErr: env.ErrOutOfRange},
		{name: "DB_HOST ausente", envName: "DB_HOST", unset: true, wantErr: env.ErrRequired},
		{name: "DB_PORT empty", envName: "DB_PORT", envValue: "", wantErr: env.ErrEmpty},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initEnv(t)

			if tt.unset {
				os.Unsetenv(tt.envName)
			} else {
				t.Setenv(tt.envName, tt.envValue)
			}

			got, err := Load()
			if !reflect.DeepEqual(got, Config{}) {
				t.Errorf("Load() = %+v; want empty Config", got)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Load() error = %v; want %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadReposrtAllErrors(t *testing.T) {
	initEnv(t)

	missing := []string{"DB_HOST", "DB_USER"}
	for _, name := range missing {
		os.Unsetenv(name)
	}

	_, err := Load()
	if err == nil {
		t.Fatalf("Load() error = nil; want errors for %v", missing)
	}

	msg := err.Error()
	for _, name := range missing {
		if !strings.Contains(msg, name) {
			t.Errorf("Load() error = %q; want mention of %s", msg, name)
		}
	}
}
