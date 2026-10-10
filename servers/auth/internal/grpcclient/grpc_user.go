package grpcclient

import (
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrGrpcConnection = errors.New("invalid URL")
)

func NewGrpcUserClient(url string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		e := fmt.Errorf("%w: %w", ErrGrpcConnection, err)
		return nil, e
	}

	return conn, nil
}
