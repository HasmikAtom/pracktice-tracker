package main

import (
	"context"

	api "github.com/HasmikAtom/tracker/v1"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Database struct {
	conn *pgxpool.Pool
}

func newDatabase(conn *pgxpool.Pool) *Database {
	return &Database{
		conn,
	}
}

type ApiServer struct {
	db *Database
}

func newApiServer(db *Database) *ApiServer {
	return &ApiServer{
		db,
	}
}

func (s *ApiServer) CreateAccount(ctx context.Context, req *api.CreateAccountRequest) (*api.CreateAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *ApiServer) GetUser(ctx context.Context, req *empty.Empty) (*api.GetUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}
