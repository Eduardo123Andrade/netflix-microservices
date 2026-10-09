package usecase

import (
	"auth/internal/entity"
	"auth/internal/repository"
	"auth/internal/utils"
	"context"
	"errors"
	"testing"
)

const fakeUserID = "01a11ce9-8e01-736d-9c09-a76d396bbc2f"

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

// fakeHasher substitui o bcrypt: é instantâneo e devolve um "hash"
// previsível, para o teste conferir o que foi gravado.
type fakeHasher struct {
	err error

	called   bool
	password string
}

func (f *fakeHasher) Hash(password string) (string, error) {
	f.called = true
	f.password = password
	if f.err != nil {
		return "", f.err
	}
	return "hashed:" + password, nil
}

// fakeUserService substitui o serviço users: devolve o userID e os erros
// configurados e registra as chamadas, inclusive a compensação.
type fakeUserService struct {
	userID    string
	createErr error
	deleteErr error

	createCalled  bool
	createName    string
	createEmail   string
	deleteCalled  bool
	deletedUserID string
}

func (f *fakeUserService) CreateUser(ctx context.Context, name, email string) (string, error) {
	f.createCalled = true
	f.createName = name
	f.createEmail = email
	if f.createErr != nil {
		return "", f.createErr
	}
	return f.userID, nil
}

func (f *fakeUserService) DeleteUser(ctx context.Context, userID string) error {
	f.deleteCalled = true
	f.deletedUserID = userID
	return f.deleteErr
}

// deps monta os três falsos no caminho feliz; cada caso muda só o que testa.
func deps() (*fakeAuthRepository, *fakeHasher, *fakeUserService) {
	return &fakeAuthRepository{findErr: repository.ErrAuthNotFound},
		&fakeHasher{},
		&fakeUserService{userID: fakeUserID}
}

var input = AuthData{
	Name:     "Eduardo",
	Email:    "teste@teste.com",
	Password: "Senha@123",
}

func TestCreateAuthUseCase(t *testing.T) {
	t.Parallel()

	t.Run("Creates user and auth", func(t *testing.T) {
		t.Parallel()

		repo, hs, us := deps()
		err := NewCreateAuthUseCase(repo, hs, us).Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}

		if repo.findEmail != input.Email {
			t.Errorf("FindByEmail() email = %q, want %q", repo.findEmail, input.Email)
		}
		if hs.password != input.Password {
			t.Errorf("Hash() password = %q, want %q", hs.password, input.Password)
		}
		if us.createName != input.Name || us.createEmail != input.Email {
			t.Errorf("CreateUser() = (%q, %q), want (%q, %q)", us.createName, us.createEmail, input.Name, input.Email)
		}
		if us.deleteCalled {
			t.Error("DeleteUser() called, want no compensation on success")
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
		if want := "hashed:" + input.Password; got.PasswordHash != want {
			t.Errorf("password_hash = %q, want %q (the plain password must be hashed)", got.PasswordHash, want)
		}
		if got.UserID != fakeUserID {
			t.Errorf("user_id = %q, want %q (the ID from the users service)", got.UserID, fakeUserID)
		}
	})

	t.Run("Error invalid input", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			email    string
			password string
			wantErr  []error
		}{
			{name: "invalid email", email: "teste.teste.com", password: input.Password, wantErr: []error{entity.ErrInvalidEmail}},
			{name: "invalid password", email: input.Email, password: "123", wantErr: []error{entity.ErrInvalidPassword}},
			{name: "both invalid", email: "teste.teste.com", password: "123", wantErr: []error{entity.ErrInvalidEmail, entity.ErrInvalidPassword}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				repo, hs, us := deps()
				data := AuthData{Name: input.Name, Email: tt.email, Password: tt.password}

				err := NewCreateAuthUseCase(repo, hs, us).Execute(context.Background(), data)
				for _, want := range tt.wantErr {
					if !errors.Is(err, want) {
						t.Errorf("Execute() = error %v; want it to include %v", err, want)
					}
				}
				if repo.findCalled {
					t.Error("FindByEmail() called, want input validated before touching the database")
				}
				if us.createCalled {
					t.Error("CreateUser() called, want it skipped on invalid input")
				}
			})
		}
	})

	t.Run("Error email already exists", func(t *testing.T) {
		t.Parallel()

		repo, hs, us := deps()
		repo.findErr = nil

		err := NewCreateAuthUseCase(repo, hs, us).Execute(context.Background(), input)
		if !errors.Is(err, repository.ErrAuthAlreadyExists) {
			t.Fatalf("Execute() = error %v; want %v", err, repository.ErrAuthAlreadyExists)
		}
		if us.createCalled {
			t.Error("CreateUser() called, want no user created for a taken email")
		}
	})

	t.Run("Error finding email is not treated as free email", func(t *testing.T) {
		t.Parallel()

		repo, hs, us := deps()
		dbErr := errors.New("connection refused")
		repo.findErr = dbErr

		err := NewCreateAuthUseCase(repo, hs, us).Execute(context.Background(), input)
		if !errors.Is(err, dbErr) {
			t.Fatalf("Execute() = error %v; want it to wrap %v", err, dbErr)
		}
		if errors.Is(err, repository.ErrAuthAlreadyExists) {
			t.Errorf("Execute() = %v; want an error other than already exists", err)
		}
		if us.createCalled {
			t.Error("CreateUser() called, want it skipped when FindByEmail fails")
		}
	})

	t.Run("Error hashing password", func(t *testing.T) {
		t.Parallel()

		repo, hs, us := deps()
		hashErr := errors.New("bcrypt: cost out of range")
		hs.err = hashErr

		err := NewCreateAuthUseCase(repo, hs, us).Execute(context.Background(), input)
		if !errors.Is(err, hashErr) {
			t.Fatalf("Execute() = error %v; want it to wrap %v", err, hashErr)
		}
		if us.createCalled {
			t.Error("CreateUser() called, want hashing done before creating the user")
		}
	})

	t.Run("Error creating user", func(t *testing.T) {
		t.Parallel()

		repo, hs, us := deps()
		usersErr := errors.New("users: unavailable")
		us.createErr = usersErr

		err := NewCreateAuthUseCase(repo, hs, us).Execute(context.Background(), input)
		if !errors.Is(err, usersErr) {
			t.Fatalf("Execute() = error %v; want it to wrap %v", err, usersErr)
		}
		if repo.createCalled {
			t.Error("CreateAuth() called, want it skipped when CreateUser fails")
		}
		if us.deleteCalled {
			t.Error("DeleteUser() called, want no compensation when no user was created")
		}
	})

	t.Run("Compensates when users returns an invalid ID", func(t *testing.T) {
		t.Parallel()

		repo, hs, us := deps()
		us.userID = "abc"

		err := NewCreateAuthUseCase(repo, hs, us).Execute(context.Background(), input)
		if !errors.Is(err, entity.ErrInvalidUserID) {
			t.Fatalf("Execute() = error %v; want it to include %v", err, entity.ErrInvalidUserID)
		}
		if repo.createCalled {
			t.Error("CreateAuth() called, want it skipped with an invalid user ID")
		}
		if us.deletedUserID != "abc" {
			t.Errorf("DeleteUser() user_id = %q, want %q", us.deletedUserID, "abc")
		}
	})

	t.Run("Compensates when saving fails", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name      string
			createErr error
		}{
			{name: "database error", createErr: errors.New("connection refused")},
			{name: "email taken between check and save", createErr: repository.ErrAuthAlreadyExists},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				repo, hs, us := deps()
				repo.createErr = tt.createErr

				err := NewCreateAuthUseCase(repo, hs, us).Execute(context.Background(), input)
				if !errors.Is(err, tt.createErr) {
					t.Fatalf("Execute() = error %v; want it to wrap %v", err, tt.createErr)
				}
				if us.deletedUserID != fakeUserID {
					t.Errorf("DeleteUser() user_id = %q, want %q", us.deletedUserID, fakeUserID)
				}
			})
		}
	})

	t.Run("Compensation failure keeps both errors", func(t *testing.T) {
		t.Parallel()

		repo, hs, us := deps()
		saveErr := errors.New("connection refused")
		deleteErr := errors.New("users: unavailable")
		repo.createErr = saveErr
		us.deleteErr = deleteErr

		err := NewCreateAuthUseCase(repo, hs, us).Execute(context.Background(), input)
		if !errors.Is(err, saveErr) {
			t.Errorf("Execute() = error %v; want it to wrap the save error %v", err, saveErr)
		}
		if !errors.Is(err, deleteErr) {
			t.Errorf("Execute() = error %v; want it to wrap the compensation error %v", err, deleteErr)
		}
	})
}
