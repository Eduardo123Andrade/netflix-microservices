package services

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jws"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

const (
	testIssuer = "auth-test"
	testTTL    = 15 * time.Minute
)

// newTestKey gera uma chave RSA em memória: os testes não dependem de arquivo
// nem de chave versionada.
func newTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey() unexpected error: %v", err)
	}
	return key
}

func newTestTokenService(t *testing.T, key *rsa.PrivateKey, ttl time.Duration) *TokenService {
	t.Helper()

	ts, err := NewTokenService(key, testIssuer, ttl)
	if err != nil {
		t.Fatalf("NewTokenService() unexpected error: %v", err)
	}
	return ts
}

// parseWithJWKS valida o token como o gateway vai fazer: só com o JWKS
// publicado, sem acesso à chave privada.
func parseWithJWKS(t *testing.T, ts *TokenService, signed string) (jwt.Token, error) {
	t.Helper()

	set, err := ts.JWKS()
	if err != nil {
		t.Fatalf("JWKS() unexpected error: %v", err)
	}
	return jwt.ParseString(signed, jwt.WithKeySet(set), jwt.WithIssuer(testIssuer))
}

func TestTokenService(t *testing.T) {
	t.Parallel()

	// Gerar chave RSA é lento; os subtestes compartilham uma só.
	key := newTestKey(t)

	const userID = "0192f5c4-7d3e-7a1b-9c2d-3e4f5a6b7c8d"

	t.Run("Token validates with the published JWKS", func(t *testing.T) {
		t.Parallel()

		ts := newTestTokenService(t, key, testTTL)

		signed, err := ts.Generate(userID)
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		tok, err := parseWithJWKS(t, ts, signed)
		if err != nil {
			t.Fatalf("parse with JWKS: %v; want a valid token", err)
		}

		if sub, _ := tok.Subject(); sub != userID {
			t.Errorf("sub = %q; want %q", sub, userID)
		}
		if iss, _ := tok.Issuer(); iss != testIssuer {
			t.Errorf("iss = %q; want %q", iss, testIssuer)
		}

		iat, ok := tok.IssuedAt()
		if !ok {
			t.Fatal("iat missing")
		}
		exp, ok := tok.Expiration()
		if !ok {
			t.Fatal("exp missing")
		}
		if got := exp.Sub(iat); got != testTTL {
			t.Errorf("exp - iat = %v; want %v", got, testTTL)
		}
	})

	t.Run("Token header carries the JWKS key id", func(t *testing.T) {
		t.Parallel()

		ts := newTestTokenService(t, key, testTTL)

		signed, err := ts.Generate(userID)
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		msg, err := jws.Parse([]byte(signed))
		if err != nil {
			t.Fatalf("jws.Parse() unexpected error: %v", err)
		}
		headers := msg.Signatures()[0].ProtectedHeaders()

		kid, ok := headers.KeyID()
		if !ok || kid == "" {
			t.Fatal("token header has no kid; the gateway cannot pick the key")
		}

		alg, _ := headers.Algorithm()
		if alg.String() != "RS256" {
			t.Errorf("alg = %q; want RS256", alg)
		}

		set, err := ts.JWKS()
		if err != nil {
			t.Fatalf("JWKS() unexpected error: %v", err)
		}
		if _, found := set.LookupKeyID(kid); !found {
			t.Errorf("kid %q not found in JWKS", kid)
		}
	})

	t.Run("JWKS exposes only the public part of the key", func(t *testing.T) {
		t.Parallel()

		ts := newTestTokenService(t, key, testTTL)

		set, err := ts.JWKS()
		if err != nil {
			t.Fatalf("JWKS() unexpected error: %v", err)
		}

		raw, err := json.Marshal(set)
		if err != nil {
			t.Fatalf("json.Marshal(JWKS) unexpected error: %v", err)
		}

		var doc struct {
			Keys []map[string]any `json:"keys"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatalf("json.Unmarshal(JWKS) unexpected error: %v", err)
		}
		if len(doc.Keys) != 1 {
			t.Fatalf("JWKS has %d keys; want 1", len(doc.Keys))
		}

		k := doc.Keys[0]
		for _, private := range []string{"d", "p", "q", "dp", "dq", "qi"} {
			if _, leaked := k[private]; leaked {
				t.Errorf("JWKS leaks private field %q", private)
			}
		}
		for _, public := range []string{"kty", "kid", "alg", "n", "e"} {
			if _, ok := k[public]; !ok {
				t.Errorf("JWKS key missing field %q", public)
			}
		}
	})

	t.Run("Token signed by another key is rejected", func(t *testing.T) {
		t.Parallel()

		ts := newTestTokenService(t, key, testTTL)
		attacker := newTestTokenService(t, newTestKey(t), testTTL)

		forged, err := attacker.Generate(userID)
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		if _, err := parseWithJWKS(t, ts, forged); err == nil {
			t.Error("token signed by another key was accepted; want error")
		}
	})

	t.Run("Tampered payload is rejected", func(t *testing.T) {
		t.Parallel()

		ts := newTestTokenService(t, key, testTTL)

		signed, err := ts.Generate(userID)
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}
		other, err := ts.Generate("0192f5c4-0000-7000-8000-000000000000")
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		// Troca o payload pelo de outro token e mantém a assinatura original:
		// é a tentativa de "virar outro usuário" editando o sub.
		parts := strings.Split(signed, ".")
		parts[1] = strings.Split(other, ".")[1]
		tampered := strings.Join(parts, ".")

		if _, err := parseWithJWKS(t, ts, tampered); err == nil {
			t.Error("tampered token was accepted; want error")
		}
	})

	t.Run("Expired token is rejected", func(t *testing.T) {
		t.Parallel()

		// TTL negativo emite um token que já nasce vencido.
		ts := newTestTokenService(t, key, -time.Minute)

		signed, err := ts.Generate(userID)
		if err != nil {
			t.Fatalf("Generate() unexpected error: %v", err)
		}

		if _, err := parseWithJWKS(t, ts, signed); err == nil {
			t.Error("expired token was accepted; want error")
		}
	})

	t.Run("Rejects nil key", func(t *testing.T) {
		t.Parallel()

		ts, err := NewTokenService(nil, testIssuer, testTTL)
		if !errors.Is(err, ErrInvalidPrivateKey) {
			t.Errorf("NewTokenService(nil) error = %v; want ErrInvalidPrivateKey", err)
		}
		if ts != nil {
			t.Error("NewTokenService(nil) returned a service; want nil")
		}
	})

	t.Run("Rejects key smaller than 2048 bits", func(t *testing.T) {
		t.Parallel()

		weak, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			t.Fatalf("rsa.GenerateKey(1024) unexpected error: %v", err)
		}

		ts, err := NewTokenService(weak, testIssuer, testTTL)
		if !errors.Is(err, ErrInvalidPrivateKey) {
			t.Errorf("NewTokenService(1024-bit key) error = %v; want ErrInvalidPrivateKey", err)
		}
		if ts != nil {
			t.Error("NewTokenService(1024-bit key) returned a service; want nil")
		}
	})

	t.Run("Key id follows the key", func(t *testing.T) {
		t.Parallel()

		a := newTestTokenService(t, key, testTTL)
		b := newTestTokenService(t, key, testTTL)
		c := newTestTokenService(t, newTestKey(t), testTTL)

		kidA, _ := a.publicKey.KeyID()
		kidB, _ := b.publicKey.KeyID()
		kidC, _ := c.publicKey.KeyID()

		// Mesma chave, mesmo kid (reiniciar o serviço não invalida tokens);
		// chave nova, kid novo (é o que permite a rotação).
		if kidA != kidB {
			t.Errorf("same key gave different kids: %q and %q", kidA, kidB)
		}
		if kidA == kidC {
			t.Errorf("different keys gave the same kid %q", kidA)
		}
	})
}

// writePEM grava um bloco PEM num diretório temporário do teste.
func writePEM(t *testing.T, blockType string, der []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "key.pem")
	data := pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der})
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("os.WriteFile() unexpected error: %v", err)
	}
	return path
}

func TestLoadRSAPrivateKey(t *testing.T) {
	t.Parallel()

	key := newTestKey(t)

	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("x509.MarshalPKCS8PrivateKey() unexpected error: %v", err)
	}

	t.Run("Loads PKCS#8 key", func(t *testing.T) {
		t.Parallel()

		got, err := LoadRSAPrivateKey(writePEM(t, "PRIVATE KEY", pkcs8))
		if err != nil {
			t.Fatalf("LoadRSAPrivateKey() unexpected error: %v", err)
		}
		if !got.Equal(key) {
			t.Error("loaded key differs from the written key")
		}
	})

	t.Run("Loads PKCS#1 key", func(t *testing.T) {
		t.Parallel()

		got, err := LoadRSAPrivateKey(writePEM(t, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key)))
		if err != nil {
			t.Fatalf("LoadRSAPrivateKey() unexpected error: %v", err)
		}
		if !got.Equal(key) {
			t.Error("loaded key differs from the written key")
		}
	})

	t.Run("Missing file", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "missing.pem")

		_, err := LoadRSAPrivateKey(path)
		if !errors.Is(err, ErrInvalidPrivateKey) {
			t.Errorf("LoadRSAPrivateKey() = %v; want ErrInvalidPrivateKey", err)
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("LoadRSAPrivateKey() = %v; want it to wrap os.ErrNotExist", err)
		}
	})

	t.Run("File is not PEM", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "key.pem")
		if err := os.WriteFile(path, []byte("not a key"), 0o600); err != nil {
			t.Fatalf("os.WriteFile() unexpected error: %v", err)
		}

		if _, err := LoadRSAPrivateKey(path); !errors.Is(err, ErrInvalidPrivateKey) {
			t.Errorf("LoadRSAPrivateKey() = %v; want ErrInvalidPrivateKey", err)
		}
	})

	t.Run("PEM with garbage bytes", func(t *testing.T) {
		t.Parallel()

		path := writePEM(t, "PRIVATE KEY", []byte("garbage"))

		if _, err := LoadRSAPrivateKey(path); !errors.Is(err, ErrInvalidPrivateKey) {
			t.Errorf("LoadRSAPrivateKey() = %v; want ErrInvalidPrivateKey", err)
		}
	})

	t.Run("Rejects non-RSA key", func(t *testing.T) {
		t.Parallel()

		ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("ecdsa.GenerateKey() unexpected error: %v", err)
		}
		der, err := x509.MarshalPKCS8PrivateKey(ec)
		if err != nil {
			t.Fatalf("x509.MarshalPKCS8PrivateKey() unexpected error: %v", err)
		}

		if _, err := LoadRSAPrivateKey(writePEM(t, "PRIVATE KEY", der)); !errors.Is(err, ErrInvalidPrivateKey) {
			t.Errorf("LoadRSAPrivateKey() = %v; want ErrInvalidPrivateKey", err)
		}
	})
}
