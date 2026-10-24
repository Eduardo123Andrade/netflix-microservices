package services

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

var (
	ErrInvalidPrivateKey = errors.New("invalid jwt private key")
	ErrNilPrivateKey     = errors.New("private key is nil")
	ErrKeyTooSmall       = errors.New("private key too small")
)

const minimumKeySizeInBits = 2048

// TokenService emite access tokens JWT assinados com RS256. A chave privada
// fica só aqui; a pública é publicada no JWKS para o gateway validar os
// tokens sem chamar o auth.
type TokenService struct {
	privateKey jwk.Key
	publicKey  jwk.Key
	issuer     string
	ttl        time.Duration
}

// NewTokenService recebe a chave já carregada (o main lê o arquivo; os testes
// podem gerar uma em memória). O "kid" é o thumbprint da chave, então muda
// sozinho quando a chave muda, o que permite rotação.
func NewTokenService(privateKey *rsa.PrivateKey, issuer string, ttl time.Duration) (*TokenService, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("token service: %w: %w", ErrNilPrivateKey, ErrInvalidPrivateKey)
	}

	if privateKey.N.BitLen() < minimumKeySizeInBits {
		return nil, fmt.Errorf("token service: %w: got %d bits, want at least %d bits: %w", ErrKeyTooSmall, privateKey.N.BitLen(), minimumKeySizeInBits, ErrInvalidPrivateKey)
	}

	priv, err := jwk.Import(privateKey)

	if err != nil {
		return nil, fmt.Errorf("token service: import private key: %w", err)
	}
	if err := jwk.AssignKeyID(priv); err != nil {
		return nil, fmt.Errorf("token service: assign key id: %w", err)
	}
	if err := priv.Set(jwk.AlgorithmKey, jwa.RS256()); err != nil {
		return nil, fmt.Errorf("token service: set algorithm: %w", err)
	}

	// A pública herda o kid e o alg da privada.
	pub, err := jwk.PublicKeyOf(priv)
	if err != nil {
		return nil, fmt.Errorf("token service: derive public key: %w", err)
	}

	return &TokenService{privateKey: priv, publicKey: pub, issuer: issuer, ttl: ttl}, nil
}

// Generate emite um access token para o usuário. O "sub" é o userID, o ID
// que pertence ao serviço users.
func (ts *TokenService) Generate(userID string) (string, error) {
	now := time.Now()

	tok, err := jwt.NewBuilder().
		Subject(userID).
		Issuer(ts.issuer).
		IssuedAt(now).
		Expiration(now.Add(ts.ttl)).
		Build()
	if err != nil {
		return "", fmt.Errorf("token service: build token: %w", err)
	}

	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256(), ts.privateKey))
	if err != nil {
		return "", fmt.Errorf("token service: sign token: %w", err)
	}

	return string(signed), nil
}

// JWKS devolve o conjunto de chaves públicas, para o endpoint
// /.well-known/jwks.json.
func (ts *TokenService) JWKS() (jwk.Set, error) {
	set := jwk.NewSet()
	if err := set.AddKey(ts.publicKey); err != nil {
		return nil, fmt.Errorf("token service: build jwks: %w", err)
	}
	return set, nil
}

// LoadRSAPrivateKey lê uma chave RSA em PEM, no formato PKCS#8 que o
// "openssl genpkey" gera (ou PKCS#1, o formato antigo).
func LoadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: read %s: %w", ErrInvalidPrivateKey, path, err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("%w: %s is not PEM", ErrInvalidPrivateKey, path)
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: parse %s: %w", ErrInvalidPrivateKey, path, err)
	}

	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("%w: %s is not an RSA key", ErrInvalidPrivateKey, path)
	}

	return key, nil
}
