package usecase

import (
	"context"
	"errors"
	"testing"

	"auth/internal/entity"
	"auth/internal/repository"
	"auth/internal/services"

	"golang.org/x/crypto/bcrypt"
)

const storedHash = "$2a$10$storedhashstoredhashstoredhashstoredhashstoredhashst"

// fakeAuthFinder substitui o repositório no login: devolve a credencial ou o
// erro configurado e registra o email buscado.
type fakeAuthFinder struct {
	auth entity.Auth
	err  error

	called bool
	email  string
}

func (f *fakeAuthFinder) FindByEmail(ctx context.Context, email string) (entity.Auth, error) {
	f.called = true
	f.email = email
	if f.err != nil {
		return entity.Auth{}, f.err
	}
	return f.auth, nil
}

// dummyHash é o que o fakeComparer devolve no Hash: formato bcrypt válido e
// diferente do storedHash, para o teste saber qual dos dois foi comparado.
const dummyHash = "$2a$10$dummyhashdummyhashdummyhashdummyhashdummyhashdummyha"

// fakeComparer substitui o bcrypt no login. O Compare devolve o erro
// configurado e registra com que hash e senha foi chamado; o Hash (usado só
// no construtor, para o hash falso) tem erro e registro próprios, para não
// se misturar com o Compare.
type fakeComparer struct {
	err     error
	hashErr error

	called   bool
	hash     string
	password string

	hashCalls int
}

func (f *fakeComparer) Compare(hash, password string) error {
	f.called = true
	f.hash = hash
	f.password = password
	return f.err
}

func (f *fakeComparer) Hash(password string) (string, error) {
	f.hashCalls++
	if f.hashErr != nil {
		return "", f.hashErr
	}
	return dummyHash, nil
}

// newLogin monta o usecase e falha o teste se o construtor der erro.
func newLogin(t *testing.T, repo authFinder, hs hashService) *LoginUseCase {
	t.Helper()

	uc, err := NewLoginUseCase(repo, hs)
	if err != nil {
		t.Fatalf("NewLoginUseCase() unexpected error: %v", err)
	}
	return uc
}

// loginDeps monta os falsos no caminho feliz: a credencial existe e a senha bate.
func loginDeps() (*fakeAuthFinder, *fakeComparer) {
	return &fakeAuthFinder{auth: entity.Auth{
			ID:           "01a11c10-8e01-736d-9c09-a76d396bbc2f",
			Email:        loginInput.Email,
			PasswordHash: storedHash,
			UserID:       fakeUserID,
		}},
		&fakeComparer{}
}

var loginInput = LoginData{
	Email:    "teste@teste.com",
	Password: "Senha@123",
}

func TestNewLoginUseCase(t *testing.T) {
	t.Parallel()

	t.Run("Fails when the dummy hash cannot be generated", func(t *testing.T) {
		t.Parallel()

		// Sem hash falso, a proteção contra ataque por tempo desligaria em
		// silêncio: melhor o serviço não subir.
		repo, hs := loginDeps()
		hashErr := errors.New("bcrypt: cost out of range")
		hs.hashErr = hashErr

		uc, err := NewLoginUseCase(repo, hs)
		if !errors.Is(err, hashErr) {
			t.Errorf("NewLoginUseCase() = error %v; want it to wrap %v", err, hashErr)
		}
		if uc != nil {
			t.Errorf("NewLoginUseCase() = %+v; want nil on error", uc)
		}
	})
}

func TestLoginUseCase(t *testing.T) {
	t.Parallel()

	t.Run("Returns the user ID when credentials match", func(t *testing.T) {
		t.Parallel()

		repo, hs := loginDeps()
		userID, err := newLogin(t, repo, hs).Execute(context.Background(), loginInput)
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}

		if userID != fakeUserID {
			t.Errorf("Execute() userID = %q, want %q (the user ID stored with the credential)", userID, fakeUserID)
		}
		if repo.email != loginInput.Email {
			t.Errorf("FindByEmail() email = %q, want %q", repo.email, loginInput.Email)
		}
		if hs.hash != storedHash || hs.password != loginInput.Password {
			t.Errorf("Compare() = (%q, %q), want (%q, %q)", hs.hash, hs.password, storedHash, loginInput.Password)
		}
	})

	t.Run("Normalizes email before searching", func(t *testing.T) {
		t.Parallel()

		repo, hs := loginDeps()
		data := LoginData{Email: "  Teste@Teste.COM ", Password: loginInput.Password}

		if _, err := newLogin(t, repo, hs).Execute(context.Background(), data); err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}

		// Cadastro grava normalizado; buscar sem normalizar não encontraria a conta.
		if repo.email != "teste@teste.com" {
			t.Errorf("FindByEmail() email = %q, want %q", repo.email, "teste@teste.com")
		}
	})

	t.Run("Error invalid email format", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			email string
		}{
			{name: "without at", email: "teste.teste.com"},
			{name: "with display name", email: "Teste <teste@teste.com>"},
			{name: "empty", email: ""},
			{name: "only spaces", email: "   "},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				repo, hs := loginDeps()
				data := LoginData{Email: tt.email, Password: loginInput.Password}

				userID, err := newLogin(t, repo, hs).Execute(context.Background(), data)
				wantNoUserID(t, userID)
				if !errors.Is(err, entity.ErrInvalidEmail) {
					t.Errorf("Execute() = error %v; want it to include %v", err, entity.ErrInvalidEmail)
				}
				if repo.called {
					t.Error("FindByEmail() called, want email validated before touching the database")
				}
			})
		}
	})

	t.Run("Does not validate password policy", func(t *testing.T) {
		t.Parallel()

		// Uma senha fora da política atual pode ser de uma conta antiga: quem
		// decide se ela vale é o hash, não a regra de cadastro.
		repo, hs := loginDeps()
		data := LoginData{Email: loginInput.Email, Password: "123"}

		userID, err := newLogin(t, repo, hs).Execute(context.Background(), data)
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if userID != fakeUserID {
			t.Errorf("Execute() userID = %q, want %q", userID, fakeUserID)
		}
	})

	t.Run("Error wrong password", func(t *testing.T) {
		t.Parallel()

		repo, hs := loginDeps()
		hs.err = services.ErrMismatchedHashAndPassword

		userID, err := newLogin(t, repo, hs).Execute(context.Background(), loginInput)
		wantNoUserID(t, userID)
		if err != ErrInvalidCredentials {
			t.Errorf("Execute() = error %v; want exactly %v", err, ErrInvalidCredentials)
		}
	})

	t.Run("Error email not found", func(t *testing.T) {
		t.Parallel()

		repo, hs := loginDeps()
		repo.err = repository.ErrAuthNotFound

		userID, err := newLogin(t, repo, hs).Execute(context.Background(), loginInput)
		wantNoUserID(t, userID)
		// Igual ao caso da senha errada: o cliente não pode descobrir quais emails têm conta.
		if err != ErrInvalidCredentials {
			t.Errorf("Execute() = error %v; want exactly %v", err, ErrInvalidCredentials)
		}
	})

	t.Run("Compares a real hash even when email is not found", func(t *testing.T) {
		t.Parallel()

		repo, hs := loginDeps()
		repo.err = repository.ErrAuthNotFound

		if _, err := newLogin(t, repo, hs).Execute(context.Background(), loginInput); err != ErrInvalidCredentials {
			t.Fatalf("Execute() = error %v; want exactly %v", err, ErrInvalidCredentials)
		}

		// Sem isso, email inexistente responde em ~1ms e existente em ~50ms+:
		// o tempo de resposta revelaria quais emails têm conta.
		if !hs.called {
			t.Fatal("Compare() not called, want bcrypt to run even for an unknown email")
		}
		if hs.password != loginInput.Password {
			t.Errorf("Compare() password = %q, want %q", hs.password, loginInput.Password)
		}
		// Um hash vazio ou inválido faz o bcrypt falhar na hora, sem gastar o tempo
		// de uma comparação de verdade.
		if _, err := bcrypt.Cost([]byte(hs.hash)); err != nil {
			t.Errorf("Compare() hash = %q, want a valid bcrypt hash: %v", hs.hash, err)
		}
		// O hash falso vem do próprio hasher (mesmo custo dos hashes salvos).
		if hs.hash != dummyHash {
			t.Errorf("Compare() hash = %q, want the dummy hash %q made by the hasher", hs.hash, dummyHash)
		}
	})

	t.Run("Generates the dummy hash once, not per login", func(t *testing.T) {
		t.Parallel()

		repo, hs := loginDeps()
		uc := newLogin(t, repo, hs)

		repo.err = repository.ErrAuthNotFound
		for range 3 {
			_, _ = uc.Execute(context.Background(), loginInput)
		}

		// Cada Hash é um bcrypt inteiro: gerar por requisição dobraria o custo do login.
		if hs.hashCalls != 1 {
			t.Errorf("Hash() called %d times, want 1 (only in the constructor)", hs.hashCalls)
		}
	})

	t.Run("Email not found stays invalid even if the dummy hash matches", func(t *testing.T) {
		t.Parallel()

		// O resultado da comparação falsa precisa ser ignorado: sem conta, não há login.
		repo, hs := loginDeps()
		repo.err = repository.ErrAuthNotFound
		hs.err = nil

		userID, err := newLogin(t, repo, hs).Execute(context.Background(), loginInput)
		wantNoUserID(t, userID)
		if err != ErrInvalidCredentials {
			t.Errorf("Execute() = error %v; want exactly %v", err, ErrInvalidCredentials)
		}
	})

	t.Run("Error corrupted hash", func(t *testing.T) {
		t.Parallel()

		repo, hs := loginDeps()
		hs.err = bcrypt.ErrHashTooShort

		userID, err := newLogin(t, repo, hs).Execute(context.Background(), loginInput)
		wantNoUserID(t, userID)

		// Para o cliente é o mesmo "não foi possível autenticar"...
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("Execute() = error %v; want it to include %v", err, ErrInvalidCredentials)
		}
		// ...mas é bug do servidor: a causa precisa seguir junto para aparecer no log.
		if !errors.Is(err, bcrypt.ErrHashTooShort) {
			t.Errorf("Execute() = error %v; want it to wrap %v", err, bcrypt.ErrHashTooShort)
		}
		if err == ErrInvalidCredentials {
			t.Error("Execute() = bare ErrInvalidCredentials, want the cause kept so it can be told apart from a wrong password")
		}
	})

	t.Run("Error finding email is not treated as invalid credentials", func(t *testing.T) {
		t.Parallel()

		repo, hs := loginDeps()
		dbErr := errors.New("connection refused")
		repo.err = dbErr

		userID, err := newLogin(t, repo, hs).Execute(context.Background(), loginInput)
		wantNoUserID(t, userID)
		if !errors.Is(err, dbErr) {
			t.Errorf("Execute() = error %v; want it to wrap %v", err, dbErr)
		}
		// Banco fora não é culpa do usuário: vira 500, não 401.
		if errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("Execute() = %v; want an error other than invalid credentials", err)
		}
	})
}
