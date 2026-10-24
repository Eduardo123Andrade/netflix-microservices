package services

import (
	"auth/internal/testutil"
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserService(t *testing.T) {
	t.Parallel()

	t.Run("CreateUser sends name and email and returns the ID", func(t *testing.T) {
		t.Parallel()

		fake := &testutil.FakeUsersServer{}
		us := NewUserService(testutil.DialUsers(t, testutil.StartFakeUsers(t, fake)))

		got, err := us.CreateUser(context.Background(), "Eduardo", "teste@teste.com")
		if err != nil {
			t.Fatalf("CreateUser() unexpected error: %v", err)
		}
		if got != testutil.FakeUserID {
			t.Errorf("CreateUser() = %q; want %q", got, testutil.FakeUserID)
		}

		reqs := fake.Created()
		if len(reqs) != 1 {
			t.Fatalf("server received %d CreateUser calls; want 1", len(reqs))
		}
		if reqs[0].GetName() != "Eduardo" || reqs[0].GetEmail() != "teste@teste.com" {
			t.Errorf("request = (%q, %q); want (%q, %q)", reqs[0].GetName(), reqs[0].GetEmail(), "Eduardo", "teste@teste.com")
		}
	})

	t.Run("CreateUser keeps the gRPC status code on error", func(t *testing.T) {
		t.Parallel()

		fake := &testutil.FakeUsersServer{CreateErr: status.Error(codes.Unavailable, "users is down")}
		us := NewUserService(testutil.DialUsers(t, testutil.StartFakeUsers(t, fake)))

		got, err := us.CreateUser(context.Background(), "Eduardo", "teste@teste.com")
		if err == nil {
			t.Fatal("CreateUser() = nil error; want error")
		}
		if got != "" {
			t.Errorf("CreateUser() = %q; want empty ID on error", got)
		}
		if code := status.Code(err); code != codes.Unavailable {
			t.Errorf("status code = %v; want %v", code, codes.Unavailable)
		}
	})

	t.Run("DeleteUser sends the ID", func(t *testing.T) {
		t.Parallel()

		fake := &testutil.FakeUsersServer{}
		us := NewUserService(testutil.DialUsers(t, testutil.StartFakeUsers(t, fake)))

		if err := us.DeleteUser(context.Background(), testutil.FakeUserID); err != nil {
			t.Fatalf("DeleteUser() unexpected error: %v", err)
		}

		deleted := fake.Deleted()
		if len(deleted) != 1 || deleted[0] != testutil.FakeUserID {
			t.Errorf("server received DeleteUser %v; want [%q]", deleted, testutil.FakeUserID)
		}
	})

	t.Run("DeleteUser returns the server error", func(t *testing.T) {
		t.Parallel()

		fake := &testutil.FakeUsersServer{DeleteErr: status.Error(codes.Internal, "boom")}
		us := NewUserService(testutil.DialUsers(t, testutil.StartFakeUsers(t, fake)))

		err := us.DeleteUser(context.Background(), testutil.FakeUserID)
		if code := status.Code(err); code != codes.Internal {
			t.Errorf("status code = %v; want %v (err %v)", code, codes.Internal, err)
		}
	})

	t.Run("Canceled context does not reach the server", func(t *testing.T) {
		t.Parallel()

		fake := &testutil.FakeUsersServer{}
		us := NewUserService(testutil.DialUsers(t, testutil.StartFakeUsers(t, fake)))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := us.CreateUser(ctx, "Eduardo", "teste@teste.com")
		if code := status.Code(err); code != codes.Canceled {
			t.Errorf("status code = %v; want %v (err %v)", code, codes.Canceled, err)
		}
		if n := len(fake.Created()); n != 0 {
			t.Errorf("server received %d CreateUser calls; want 0", n)
		}
	})
}
