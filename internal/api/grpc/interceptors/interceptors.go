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
func NewAuthHandler(sec secretary.Secretary, cfg *config.Config) (*AuthHandler, error) {
	return &AuthHandler{
		sec: sec,
		cfg: cfg,
	}, nil
}

// AuthFunc is the pluggable function that performs authentication.
//
// The passed in `Context` will contain the gRPC metadata.MD object (for header-based authentication) and
// the peer.Peer information that can contain transport-based credentials (e.g. `credentials.AuthInfo`).
//
// The returned context will be propagated to handlers, allowing user changes to `Context`. However,
// please make sure that the `Context` returned is a child `Context` of the one passed in.
//
// If error is returned, its `grpc.Code()` will be returned to the user as well as the verbatim message.
// Please make sure you use `codes.Unauthenticated` (lacking auth) and `codes.PermissionDenied`
// (authed, but lacking perms) appropriately.
func (a *AuthHandler) AuthFunc(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		userID := uuid.New().String()
		token := a.sec.Encode(userID)
		md.Set(UserAuthKey, token)
		newCtx := metadata.NewOutgoingContext(ctx, md)
		return newCtx, nil
	}
	values := md.Get(UserAuthKey)
	if len(values) <= 0 {
		userID := uuid.New().String()
		token := a.sec.Encode(userID)
		md.Set(UserAuthKey, token)
		newCtx := metadata.NewOutgoingContext(ctx, md)
		return newCtx, nil
	}
	_, err := a.sec.Decode(values[0])
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return ctx, nil
}

// UnaryServerInterceptor returns a new unary server interceptors that performs per-request auth.
func (a *AuthHandler) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx, err := a.AuthFunc(ctx)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}
