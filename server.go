package main

import (
	"context"
	"log"
	"net"
	"net/http"

	api "github.com/HasmikAtom/tracker/v1"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func serveHTTP(l net.Listener, rootMux *http.ServeMux) {
	s := &http.Server{
		Handler: rootMux,
	}
	if err := s.Serve(l); err != cmux.ErrListenerClosed {
		log.Fatalf("Error in HTTP server: %s", err)
	}
}

func serveGRPC(l net.Listener, server *ApiServer) {
	grpcs := grpc.NewServer()
	api.RegisterApiServer(grpcs, server)
	if err := grpcs.Serve(l); err != cmux.ErrListenerClosed {
		log.Fatalf("Error in GRPC server: %s", err)
	}
}

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

func (s *ApiServer) Login(ctx context.Context, req *api.LoginRequest) (*api.LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *ApiServer) GetUser(ctx context.Context, req *empty.Empty) (*api.GetUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *ApiServer) UpdateUser(ctx context.Context, req *api.UpdateUserRequest) (*api.UpdateUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *ApiServer) DeleteAccount(ctx context.Context, req *api.DeleteAccountRequest) (*api.DeleteAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *ApiServer) CreateTicket(ctx context.Context, req *api.CreateTicketRequest) (*api.CreateTicketResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *ApiServer) ListTickets(ctx context.Context, req *emptypb.Empty) (*api.ListTicketsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *ApiServer) FilterTickets(ctx context.Context, req *api.FilterTicketsRequest) (*api.FilterTicketsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *ApiServer) GetTicket(ctx context.Context, req *api.GetTicketRequest) (*api.GetTicketResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *ApiServer) UpdateTicket(ctx context.Context, req *api.UpdateTicketRequest) (*api.UpdateTicketResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *ApiServer) DeleteTicket(ctx context.Context, req *api.DeleteTicketRequest) (*api.DeleteTicketResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}
