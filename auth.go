package main

import (
	"context"
	"encoding/json"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Making this available globally makes it available for retrieving the
// user id from the context inside the endpoints too.
type userID string

var user userID = "userID"

// http auth middleware func
func httpAuthNFunc(ctx context.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if userID := r.Header.Get("x-atom-user"); userID != "" {
				ctx = context.WithValue(r.Context(), user, userID)

				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				err := status.New(codes.Unauthenticated, "please log in")
				// Formatting the http error to be similar to grpc error responses
				res := struct {
					Code    codes.Code    `json:"code"`
					Message string        `json:"message"`
					Details []interface{} `json:"details"`
				}{
					Code:    err.Code(),
					Message: err.Message(),
					Details: err.Details(),
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(res)
			}
		})
	}
}

// grpc auth middleware func
func grpcAuthNFunc(ctx context.Context) func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		c := ctx

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if userID := md.Get("X-Atom-User")[0]; userID != "" {
				c = context.WithValue(ctx, user, userID)
				h, err := handler(c, req)
				return h, err
			}
		}

		return nil, status.Errorf(codes.Unauthenticated, "please log in")
	}
}
