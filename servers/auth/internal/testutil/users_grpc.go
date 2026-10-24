package testutil

import (
	usersv1 "auth/internal/gen/users/v1"
	"auth/internal/grpcclient"
	"context"
	"net"
	"sync"
	"testing"

	"google.golang.org/grpc"
)

// FakeUserID é um UUID v7 válido, devolvido pelo falso quando o teste não
// configura outro.
const FakeUserID = "01a11ce9-8e01-736d-9c09-a76d396bbc2f"

// FakeUsersServer implementa o servidor gerado do contrato users.v1. Devolve
// o ID e os erros configurados e registra as requisições recebidas.
//
// Os campos de configuração (UserID, CreateErr, DeleteErr) devem ser
// definidos antes das chamadas; o registro é protegido por mutex porque o
// servidor atende subtestes em paralelo.
type FakeUsersServer struct {
	usersv1.UnimplementedUserServiceServer

	UserID    string
	CreateErr error
	DeleteErr error

	mu      sync.Mutex
	created []*usersv1.CreateUserRequest
	deleted []string
}

func (f *FakeUsersServer) CreateUser(ctx context.Context, req *usersv1.CreateUserRequest) (*usersv1.CreateUserResponse, error) {
	f.mu.Lock()
	f.created = append(f.created, req)
	f.mu.Unlock()

	if f.CreateErr != nil {
		return nil, f.CreateErr
	}
	return &usersv1.CreateUserResponse{UserId: f.UserID}, nil
}

func (f *FakeUsersServer) DeleteUser(ctx context.Context, req *usersv1.DeleteUserRequest) (*usersv1.DeleteUserResponse, error) {
	f.mu.Lock()
	f.deleted = append(f.deleted, req.GetUserId())
	f.mu.Unlock()

	if f.DeleteErr != nil {
		return nil, f.DeleteErr
	}
	return &usersv1.DeleteUserResponse{}, nil
}

// Created devolve uma cópia das requisições de CreateUser recebidas.
func (f *FakeUsersServer) Created() []*usersv1.CreateUserRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]*usersv1.CreateUserRequest(nil), f.created...)
}

// Deleted devolve uma cópia dos IDs recebidos em DeleteUser.
func (f *FakeUsersServer) Deleted() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.deleted...)
}

// StartFakeUsers sobe o servidor falso numa porta TCP real e livre e devolve
// o endereço ("127.0.0.1:<porta>"). A porta 0 pede ao sistema qualquer porta
// livre, então testes em paralelo e servidores locais nunca colidem. O
// servidor é encerrado no fim do teste.
func StartFakeUsers(t *testing.T, fake *FakeUsersServer) string {
	t.Helper()

	if fake.UserID == "" {
		fake.UserID = FakeUserID
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen fake users: %v", err)
	}

	srv := grpc.NewServer()
	usersv1.RegisterUserServiceServer(srv, fake)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	return lis.Addr().String()
}

// DialUsers conecta no endereço com o mesmo conector da produção e devolve o
// cliente gerado. A conexão é fechada no fim do teste.
func DialUsers(t *testing.T, addr string) usersv1.UserServiceClient {
	t.Helper()

	conn, err := grpcclient.NewGrpcUserClient(addr)
	if err != nil {
		t.Fatalf("NewGrpcUserClient(%q): %v", addr, err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return usersv1.NewUserServiceClient(conn)
}
