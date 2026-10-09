package usecase

import (
	"auth/internal/entity"
	"auth/internal/repository"
	"auth/internal/utils"
	"context"
	"errors"
	"testing"
)

// fakeAuthRepository substitui o repositório real: devolve os erros
// configurados e registra as chamadas para o teste conferir.
type fakeAuthRepository struct {
	findErr   error
	createErr error

	findCalled    bool
	findEmail     string
	createCalled  bool
	createdEntity entity.Auth
}

func (f *fakeAuthRepository) FindByEmail(ctx context.Context, email string) (entity.Auth, error) {
	f.findCalled = true
	f.findEmail = email
	if f.findErr != nil {
		return entity.Auth{}, f.findErr
	}
	return entity.Auth{ID: "01a11c10-8e01-736d-9c09-a76d396bbc2f", Email: email}, nil
}

func (f *fakeAuthRepository) CreateAuth(ctx context.Context, data entity.Auth) error {
	f.createCalled = true
	f.createdEntity = data
	return f.createErr
}

func TestCreateAuthUseCase(t *testing.T) {
	t.Parallel()

	input := AuthData{
		Email:        "teste@teste.com",
		PasswordHash: "$2a$10$fakehash",
		UserID:       "01a11ce9-8e01-736d-9c09-a76d396bbc2f",
	}

	t.Run("Creates auth when email is free", func(t *testing.T) {
		t.Parallel()

		repo := &fakeAuthRepository{findErr: repository.ErrAuthNotFound}
		uc := NewAuthUseCase(repo)

		err := uc.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}

		if repo.findEmail != input.Email {
			t.Errorf("FindByEmail() email = %q, want %q", repo.findEmail, input.Email)
		}
		if !repo.createCalled {
			t.Fatal("CreateAuth() not called, want it called")
		}

		got := repo.createdEntity
		if !utils.IsUUIDV7(got.ID) {
			t.Errorf("id = %q, want a UUID v7", got.ID)
		}
		if got.Email != input.Email {
			t.Errorf("email = %q, want %q", got.Email, input.Email)
		}
		if got.PasswordHash != input.PasswordHash {
			t.Errorf("password_hash = %q, want %q", got.PasswordHash, input.PasswordHash)
		}
		if got.UserID != input.UserID {
			t.Errorf("user_id = %q, want %q", got.UserID, input.UserID)
		}
	})

	t.Run("Error email already exists", func(t *testing.T) {
		t.Parallel()

		repo := &fakeAuthRepository{findErr: nil}
		uc := NewAuthUseCase(repo)

		err := uc.Execute(context.Background(), input)
		if !errors.Is(err, repository.ErrAuthAlreadyExists) {
			t.Fatalf("Execute() = error %v; want %v", err, repository.ErrAuthAlreadyExists)
		}
		if repo.createCalled {
			t.Error("CreateAuth() called, want it skipped when email exists")
		}
	})

	t.Run("Error finding email is not treated as free email", func(t *testing.T) {
		t.Parallel()

		dbErr := errors.New("connection refused")
		repo := &fakeAuthRepository{findErr: dbErr}
		uc := NewAuthUseCase(repo)

		err := uc.Execute(context.Background(), input)
		if !errors.Is(err, dbErr) {
			t.Fatalf("Execute() = error %v; want it to wrap %v", err, dbErr)
		}
		if errors.Is(err, repository.ErrAuthAlreadyExists) {
			t.Errorf("Execute() = %v; want an error other than already exists", err)
		}
		if repo.createCalled {
			t.Error("CreateAuth() called, want it skipped when FindByEmail fails")
		}
	})

	t.Run("Error email taken between check and save", func(t *testing.T) {
		t.Parallel()

		repo := &fakeAuthRepository{
			findErr:   repository.ErrAuthNotFound,
			createErr: repository.ErrAuthAlreadyExists,
		}
		uc := NewAuthUseCase(repo)

		err := uc.Execute(context.Background(), input)
		if !errors.Is(err, repository.ErrAuthAlreadyExists) {
			t.Fatalf("Execute() = error %v; want %v", err, repository.ErrAuthAlreadyExists)
		}
	})

	t.Run("Error saving auth", func(t *testing.T) {
		t.Parallel()

		dbErr := errors.New("connection refused")
		repo := &fakeAuthRepository{
			findErr:   repository.ErrAuthNotFound,
			createErr: dbErr,
		}
		uc := NewAuthUseCase(repo)

		err := uc.Execute(context.Background(), input)
		if !errors.Is(err, dbErr) {
			t.Fatalf("Execute() = error %v; want it to wrap %v", err, dbErr)
		}
	})

	t.Run("Error invalid input", func(t *testing.T) {
		t.Parallel()

		repo := &fakeAuthRepository{findErr: repository.ErrAuthNotFound}
		uc := NewAuthUseCase(repo)

		invalid := AuthData{
			Email:        "teste.teste.com",
			PasswordHash: input.PasswordHash,
			UserID:       "abc",
		}

		err := uc.Execute(context.Background(), invalid)
		if !errors.Is(err, entity.ErrInvalidEmail) {
			t.Errorf("Execute() = error %v; want it to include %v", err, entity.ErrInvalidEmail)
		}
		if !errors.Is(err, entity.ErrInvalidUserID) {
			t.Errorf("Execute() = error %v; want it to include %v", err, entity.ErrInvalidUserID)
		}
		if repo.findCalled {
			t.Error("FindByEmail() called, want input validated before touching the database")
		}
		if repo.createCalled {
			t.Error("CreateAuth() called, want it skipped on invalid input")
		}
	})
}
