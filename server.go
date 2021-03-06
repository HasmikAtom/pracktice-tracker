package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	api "github.com/HasmikAtom/tracker/v1"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/soheilhy/cmux"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func serveHTTP(l net.Listener, rootMux *http.ServeMux) {
	s := &http.Server{
		Handler: rootMux,
	}
	if err := s.Serve(l); err != cmux.ErrListenerClosed {
		log.Fatalf("Error in HTTP server: %s", err)
	}
}

func serveGRPC(l net.Listener, server *ApiServer, grpcAuthN func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)) {
	grpcs := grpc.NewServer(
		grpc.UnaryInterceptor(grpcAuthN),
	)
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

func genRandomBytesHex(amount uint32) string {
	buf := make([]byte, amount)
	_, err := rand.Read(buf)
	if err != nil {
		panic(fmt.Errorf("rand.Read: %v", err))
	}
	return hex.EncodeToString(buf)
}

func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 5)
	if err != nil {
		panic(fmt.Errorf("bcrypt.GenerateFromPassword: %v", err))
	}
	return string(hash)
}

func (s *ApiServer) CreateAccount(ctx context.Context, req *api.CreateAccountRequest) (*api.CreateAccountResponse, error) {
	if req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email is required")
	}
	if req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "password is required")
	}
	if req.FirstName == "" {
		return nil, status.Errorf(codes.InvalidArgument, "firstname is required")
	}
	if req.LastName == "" {
		return nil, status.Errorf(codes.InvalidArgument, "lastname is required")
	}

	tx, err := s.db.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}
	defer tx.Rollback(ctx)

	var email string
	row := tx.QueryRow(ctx, `
		SELECT email
		FROM users
		WHERE email = $1
	`, req.Email)
	err = row.Scan(&email)
	if err == nil {
		if email != "" {
			return nil, status.Errorf(codes.AlreadyExists, "user already exists")
		}
	} else {
		if err != pgx.ErrNoRows {
			return nil, status.Errorf(codes.Internal, "failed to create account")
		}
	}

	emailVerifyToken := genRandomBytesHex(20)

	row = tx.QueryRow(ctx, `
		INSERT INTO users (email, password, first_name, last_name, email_verify_token, user_type, auth_method)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, email, user_type, activated_at, email_verified, first_name, last_name, auth_method, created_at, deleted_at
	`,
		req.Email,
		hashPassword(req.Password),
		req.FirstName,
		req.LastName,
		emailVerifyToken,
		req.UserType,
		req.AuthMethod,
	)
	var (
		user         api.User
		dbAuthMethod sql.NullString
		activatedAt  sql.NullTime
		created      time.Time
		deleted      sql.NullTime
	)
	err = row.Scan(
		&user.Id,
		&user.Email,
		&user.UserType,
		&activatedAt,
		&user.EmailVerified,
		&user.FirstName,
		&user.LastName,
		&dbAuthMethod,
		&created,
		&deleted,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create account: %v", err)
	}

	if dbAuthMethod.Valid {
		user.AuthMethod = dbAuthMethod.String
	}
	if activatedAt.Valid {
		user.ActivatedAt = timestamppb.New(activatedAt.Time)
	}
	if user.CreatedAt = timestamppb.New(created.Truncate(60 * time.Second)); err != nil {
		return nil, status.Errorf(codes.Internal, "error creating account: %v", err)
	}
	if deleted.Valid {
		user.DeletedAt = timestamppb.New(deleted.Time.Truncate(60 * time.Second))
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "error creating account: %v", err)
	}

	return &api.CreateAccountResponse{User: &user, Message: fmt.Sprintf("Account for the user %s created", user.Email)}, nil
}

func (s *ApiServer) Login(ctx context.Context, req *api.LoginRequest) (*api.LoginResponse, error) {
	if req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email is required")
	}
	if req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "password is required")
	}

	tx, err := s.db.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to login: %v", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		SELECT id, email, password, user_type, activated_at, email_verified, first_name, last_name, auth_method, created_at, deleted_at
		FROM users
		WHERE email = $1
	`, req.Email)
	var (
		user         api.User
		password     string
		dbAuthMethod sql.NullString
		activatedAt  sql.NullTime
		created      time.Time
		deleted      sql.NullTime
	)
	err = row.Scan(
		&user.Id,
		&user.Email,
		&password,
		&user.UserType,
		&activatedAt,
		&user.EmailVerified,
		&user.FirstName,
		&user.LastName,
		&dbAuthMethod,
		&created,
		&deleted,
	)
	if err != nil {
		if err != pgx.ErrNoRows {
			return nil, status.Errorf(codes.InvalidArgument, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to login")
	}

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "email and password don't match")
	}

	if dbAuthMethod.Valid {
		user.AuthMethod = dbAuthMethod.String
	}
	if activatedAt.Valid {
		user.ActivatedAt = timestamppb.New(activatedAt.Time.Truncate(60 * time.Second))
	}
	if user.CreatedAt = timestamppb.New(created.Truncate(60 * time.Second)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to login: %v", err)
	}
	if deleted.Valid {
		user.DeletedAt = timestamppb.New(deleted.Time.Truncate(60 * time.Second))
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to login: %v", err)
	}

	return &api.LoginResponse{User: &user, Message: fmt.Sprintf("user %s successfully logged in", user.Email)}, nil
}

func (s *ApiServer) GetUser(ctx context.Context, req *empty.Empty) (*api.GetUserResponse, error) {
	userID := ctx.Value(user)

	tx, err := s.db.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve user: %v", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		SELECT id, email, user_type, activated_at, email_verified, first_name, last_name, auth_method, created_at, deleted_at
		FROM users
		WHERE id = $1
	`, userID)
	var (
		user         api.User
		dbAuthMethod sql.NullString
		activatedAt  sql.NullTime
		created      time.Time
		deleted      sql.NullTime
	)
	err = row.Scan(
		&user.Id,
		&user.Email,
		&user.UserType,
		&activatedAt,
		&user.EmailVerified,
		&user.FirstName,
		&user.LastName,
		&dbAuthMethod,
		&created,
		&deleted,
	)
	if err != nil {
		if err != pgx.ErrNoRows {
			return nil, status.Errorf(codes.InvalidArgument, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve user")
	}

	if dbAuthMethod.Valid {
		user.AuthMethod = dbAuthMethod.String
	}
	if activatedAt.Valid {
		user.ActivatedAt = timestamppb.New(activatedAt.Time.Truncate(60 * time.Second))
	}
	if user.CreatedAt = timestamppb.New(created.Truncate(60 * time.Second)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve user: %v", err)
	}
	if deleted.Valid {
		user.DeletedAt = timestamppb.New(deleted.Time.Truncate(60 * time.Second))
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve user: %v", err)
	}

	return &api.GetUserResponse{User: &user, Message: fmt.Sprintf("user %s retrieved", user.Email)}, nil
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

func (s *ApiServer) CreateGroup(ctx context.Context, req *api.CreateGroupRequest) (*api.CreateGroupResponse, error) {
	userID := ctx.Value(user)
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	tx, err := s.db.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create group: %v", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		INSERT INTO groups (user_id, group_name, descript)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, group_name, descript, created_at, deleted_at
	`,
		userID,
		req.Name,
		req.Description,
	)
	var (
		group       api.Group
		description sql.NullString
		created     time.Time
		deleted     sql.NullTime
	)
	err = row.Scan(
		&group.Id,
		&group.UserId,
		&group.Name,
		&description,
		&created,
		&deleted,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create group: %v", err)
	}
	if description.Valid {
		group.Description = description.String
	}
	if group.CreatedAt = timestamppb.New(created.Truncate(60 * time.Second)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create group: %v", err)
	}
	if deleted.Valid {
		group.DeletedAt = timestamppb.New(deleted.Time.Truncate(60 * time.Second))
	}

	row = tx.QueryRow(ctx, `
		INSERT INTO group_users (group_id, user_id, permission)
		VALUES ($1, $2, $3)
	`, group.Id, group.UserId, "admin")

	err = row.Scan()
	if err != pgx.ErrNoRows {
		return nil, status.Errorf(codes.Internal, "failed to create group: %v", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create group: %v", err)
	}
	return &api.CreateGroupResponse{Group: &group, Message: fmt.Sprintf("group %s created", group.Name)}, nil
}

func (s *ApiServer) GetGroup(ctx context.Context, req *api.GetGroupRequest) (*api.GetGroupResponse, error) {
	userID := ctx.Value(user)
	if req.GroupId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "group id required")
	}

	tx, err := s.db.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve group: %v", err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		SELECT groups.id, groups.user_id, groups.group_name, groups.descript, groups.created_at, groups.deleted_at
		FROM groups
		JOIN group_users ON groups.id = group_users.group_id
		WHERE group_users.group_id = $1
		AND group_users.user_id = $2
		`,
		req.GroupId,
		userID,
	)
	var (
		group       api.Group
		description sql.NullString
		created     time.Time
		deleted     sql.NullTime
	)
	err = row.Scan(
		&group.Id,
		&group.UserId,
		&group.Name,
		&description,
		&created,
		&deleted,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve group: %v", err)
	}
	if description.Valid {
		group.Description = description.String
	}

	if group.CreatedAt = timestamppb.New(created.Truncate(60 * time.Second)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve group: %v", err)
	}
	if deleted.Valid {
		group.DeletedAt = timestamppb.New(deleted.Time.Truncate(60 * time.Second))
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve group: %v", err)
	}

	return &api.GetGroupResponse{Group: &group, Message: fmt.Sprintf("group %s retrieved", group.Name)}, nil
}

func (s *ApiServer) ListGroups(ctx context.Context, req *empty.Empty) (*api.ListGroupsResponse, error) {
	userID := ctx.Value(user)

	tx, err := s.db.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve groups: %v", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT groups.id, groups.user_id, groups.group_name, groups.descript, groups.created_at, groups.deleted_at
		FROM groups
		JOIN group_users ON groups.id = group_users.group_id
		WHERE group_users.user_id = $1
	`, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve groups: %v", err)
	}
	defer rows.Close()
	var groups []*api.Group
	for rows.Next() {

		var (
			group       api.Group
			description sql.NullString
			created     time.Time
			deleted     sql.NullTime
		)
		err = rows.Scan(
			&group.Id,
			&group.UserId,
			&group.Name,
			&description,
			&created,
			&deleted,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to retrieve groups: %v", err)
		}
		if description.Valid {
			group.Description = description.String
		}
		if group.CreatedAt = timestamppb.New(created.Truncate(60 * time.Second)); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to retrieve groups: %v", err)
		}
		if deleted.Valid {
			group.DeletedAt = timestamppb.New(deleted.Time.Truncate(60 * time.Second))
		}
		groups = append(groups, &group)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve groups: %v", err)
	}
	return &api.ListGroupsResponse{Groups: groups, Message: "groups retrieved"}, nil
}
