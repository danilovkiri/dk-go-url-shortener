// Package interceptors provides various middleware functionality for GRPC.
package interceptors

import (
	"context"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/secretary"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthHandler sets object structure.
type AuthHandler struct {
	sec secretary.Secretary
	cfg *config.Config
}

// UserAuthKey sets a user key to be used in user identification.
const UserAuthKey = "user"

// NewAuthHandler initializes a new cookie handler.
func NewAuthHandler(sec secretary.Secretary, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		sec: sec,
		cfg: cfg,
	}
}

// AuthFunc is the pluggable function that performs authentication.
func (a *AuthHandler) AuthFunc(ctx context.Context) (context.Context, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		userID := uuid.New().String()
		token := a.sec.Encode(userID)
		newMd := metadata.New(map[string]string{UserAuthKey: token})
		newCtx := metadata.NewIncomingContext(ctx, newMd)
		return newCtx, token, nil
	}
	values := md.Get(UserAuthKey)
	if len(values) == 0 {
		userID := uuid.New().String()
		token := a.sec.Encode(userID)
		newMd := metadata.New(map[string]string{UserAuthKey: token})
		newCtx := metadata.NewIncomingContext(ctx, newMd)
		return newCtx, token, nil
	}
	_, err := a.sec.Decode(values[0])
	if err != nil {
		return nil, "", status.Error(codes.PermissionDenied, err.Error())
	}
	return ctx, "", nil
}

// UnaryServerInterceptor returns a new unary server interceptors that performs per-request auth.
func (a *AuthHandler) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx, token, err := a.AuthFunc(ctx)
		if err != nil {
			return nil, err
		}
		if token != "" {
			err = grpc.SendHeader(newCtx, metadata.New(map[string]string{UserAuthKey: token}))
			if err != nil {
				return nil, err
			}
		}
		return handler(newCtx, req)
	}
}
