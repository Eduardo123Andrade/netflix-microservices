package main

import (
	"fmt"

	"auth/internal/config"
	"auth/internal/services"
)

// newTokenService carrega a chave privada e monta o emissor de tokens. A
// validação do arquivo (existe, é PEM, é RSA, tem 2048 bits ou mais) fica no
// próprio services; aqui só ligamos as peças.
func newTokenService(c config.JWT) (*services.TokenService, error) {
	key, err := services.LoadRSAPrivateKey(c.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load jwt key: %w", err)
	}

	ts, err := services.NewTokenService(key, c.Issuer, c.TTL)
	if err != nil {
		return nil, fmt.Errorf("create token service: %w", err)
	}

	return ts, nil
}
