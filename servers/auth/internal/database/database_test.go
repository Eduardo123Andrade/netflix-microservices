package database

import (
	"auth/internal/config"
	"context"
	"errors"
	"testing"
)

func TestNewPool(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Database
	}{
		{
			name: "valores comuns",
			cfg:  config.Database{User: "test_user", Password: "test_password", Name: "test_db", Host: "localhost", Port: 5432},
		},
		{
			name: "senha com caracteres especiais",
			cfg:  config.Database{User: "test_user", Password: "p@ss:w/rd?#%", Name: "test_db", Host: "localhost", Port: 5432},
		},
		{
			name: "usuário com caractere especial",
			cfg:  config.Database{User: "us@er", Password: "test_password", Name: "test_db", Host: "localhost", Port: 5432},
		},
		{
			name: "host IPv6",
			cfg:  config.Database{User: "test_user", Password: "test_password", Name: "test_db", Host: "::1", Port: 5432},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := NewPool(context.Background(), tt.cfg)
			if err != nil {
				t.Fatalf("NewPool() unexpected error: %v", err)
			}
			t.Cleanup(pool.Close)

			cc := pool.Config().ConnConfig
			got := config.Database{
				User:     cc.User,
				Password: cc.Password,
				Name:     cc.Database,
				Host:     cc.Host,
				Port:     int(cc.Port),
			}

			if got != tt.cfg {
				t.Errorf("NewPool() config = %+v; want %+v", got, tt.cfg)
			}
		})
	}
}

func TestNewPoolError(t *testing.T) {
	pool, err := NewPool(context.Background(), config.Database{
		User: "test_user", Password: "test_password", Name: "test_db", Host: "localhost", Port: -1,
	})

	if !errors.Is(err, ErrNewPool) {
		t.Errorf("NewPool() err = %+v; want = %v", err, ErrNewPool)
	}

	if pool != nil {
		t.Errorf("NewPool() pool = %+v; want = %v", pool, nil)
	}

}
