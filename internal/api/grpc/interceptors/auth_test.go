package interceptors

import (
	"context"
	"net"
	"testing"

	"github.com/danilovkiri/dk_go_url_shortener/internal/api/grpc/handlers"
	pb "github.com/danilovkiri/dk_go_url_shortener/internal/api/grpc/proto"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/mocks"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/secretary/v1"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestAuthHandler_AuthFunc_NoMD(t *testing.T) {
	cfg := config.NewDefaultConfiguration()
	cfg.UserKey = "jds__63h3_7ds"
	secretaryService := secretary.NewSecretaryService(cfg)
	authHandler := NewAuthHandler(secretaryService, cfg)
	ctx := context.Background()
	newCtx, token, err := authHandler.AuthFunc(ctx)
	assert.Equal(t, nil, err)
	md, ok := metadata.FromIncomingContext(newCtx)
	assert.Equal(t, true, ok)
	assert.Equal(t, token, md.Get(UserAuthKey)[0])
}

func TestAuthHandler_AuthFunc_EmptyMD(t *testing.T) {
	cfg := config.NewDefaultConfiguration()
	cfg.UserKey = "jds__63h3_7ds"
	secretaryService := secretary.NewSecretaryService(cfg)
	authHandler := NewAuthHandler(secretaryService, cfg)
	md := metadata.New(map[string]string{"some_key": "some_token"})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	newCtx, token, err := authHandler.AuthFunc(ctx)
	assert.Equal(t, nil, err)
	md, ok := metadata.FromIncomingContext(newCtx)
	assert.Equal(t, true, ok)
	assert.Equal(t, token, md.Get(UserAuthKey)[0])
}

func TestAuthHandler_AuthFunc_CorrectMD(t *testing.T) {
	cfg := config.NewDefaultConfiguration()
	cfg.UserKey = "jds__63h3_7ds"
	secretaryService := secretary.NewSecretaryService(cfg)
	authHandler := NewAuthHandler(secretaryService, cfg)
	token := "8773a90a68ebd0fd56dffb1441682414fbec5f454eba9be6129bb00744f50d7f19fd870e97eba101a03b857c675e4836de6f5196"
	md := metadata.New(map[string]string{"user": token})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	newCtx, _, err := authHandler.AuthFunc(ctx)
	assert.Equal(t, nil, err)
	assert.Equal(t, ctx, newCtx)
}

func TestAuthHandler_AuthFunc_IncorrectMD(t *testing.T) {
	cfg := config.NewDefaultConfiguration()
	cfg.UserKey = "jds__63h3_7ds"
	secretaryService := secretary.NewSecretaryService(cfg)
	authHandler := NewAuthHandler(secretaryService, cfg)
	token := "some_incorrect_token"
	md := metadata.New(map[string]string{"user": token})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	newCtx, _, err := authHandler.AuthFunc(ctx)
	assert.Equal(t, nil, newCtx)
	if e, ok := status.FromError(err); ok {
		assert.Equal(t, codes.PermissionDenied, e.Code())
	} else {
		t.Fatal("Error code was not retrieved")
	}
}

func TestAuthHandler_UnaryServerInterceptor_NoMD(t *testing.T) {
	cfg := config.NewDefaultConfiguration()
	cfg.UserKey = "jds__63h3_7ds"
	secretaryService := secretary.NewSecretaryService(cfg)
	authHandler := NewAuthHandler(secretaryService, cfg)

	// set up a GRPC server
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(authHandler.UnaryServerInterceptor()))
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storageInit := mocks.NewMockURLStorage(ctrl)
	storageInit.EXPECT().PingDB().Return(nil)
	server, err := handlers.InitServer(context.Background(), cfg, storageInit)
	if err != nil {
		t.Fatal(err)
	}
	pb.RegisterShortenerServer(s, server)
	go s.Serve(listen)

	// set up a GRPC client
	conn, err := grpc.Dial(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// send a request
	ctx := context.Background()
	var header, trailer metadata.MD
	c := pb.NewShortenerClient(conn)
	var request emptypb.Empty
	resp, err := c.PingDB(ctx, &request, grpc.Header(&header), grpc.Trailer(&trailer))
	assert.Equal(t, []string{"application/grpc"}, header.Get("content-type"))
	assert.NotEmpty(t, header.Get("user"))
	s.GracefulStop()
	assert.Equal(t, nil, err)
	assert.NotEmpty(t, resp)
}

func TestAuthHandler_UnaryServerInterceptor_CorrectMD(t *testing.T) {
	cfg := config.NewDefaultConfiguration()
	cfg.UserKey = "jds__63h3_7ds"
	secretaryService := secretary.NewSecretaryService(cfg)
	authHandler := NewAuthHandler(secretaryService, cfg)

	// set up a GRPC server
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(authHandler.UnaryServerInterceptor()))
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storageInit := mocks.NewMockURLStorage(ctrl)
	storageInit.EXPECT().PingDB().Return(nil)
	server, err := handlers.InitServer(context.Background(), cfg, storageInit)
	if err != nil {
		t.Fatal(err)
	}
	pb.RegisterShortenerServer(s, server)
	go s.Serve(listen)

	// set up a GRPC client
	conn, err := grpc.Dial(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// send a request
	token := "8773a90a68ebd0fd56dffb1441682414fbec5f454eba9be6129bb00744f50d7f19fd870e97eba101a03b857c675e4836de6f5196"
	md := metadata.New(map[string]string{"user": token})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	c := pb.NewShortenerClient(conn)
	var request emptypb.Empty
	resp, err := c.PingDB(ctx, &request)
	s.GracefulStop()
	assert.Equal(t, nil, err)
	assert.NotEmpty(t, resp)
}

func TestAuthHandler_UnaryServerInterceptor_IncorrectMD(t *testing.T) {
	cfg := config.NewDefaultConfiguration()
	cfg.UserKey = "jds__63h3_7ds"
	secretaryService := secretary.NewSecretaryService(cfg)
	authHandler := NewAuthHandler(secretaryService, cfg)

	// set up a GRPC server
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(authHandler.UnaryServerInterceptor()))
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	storageInit := mocks.NewMockURLStorage(ctrl)
	server, err := handlers.InitServer(context.Background(), cfg, storageInit)
	if err != nil {
		t.Fatal(err)
	}
	pb.RegisterShortenerServer(s, server)
	go s.Serve(listen)

	// set up a GRPC client
	conn, err := grpc.Dial(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// send a request
	token := "some_incorrect_token"
	md := metadata.New(map[string]string{"user": token})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	c := pb.NewShortenerClient(conn)
	var request emptypb.Empty
	_, err = c.PingDB(ctx, &request)
	s.GracefulStop()
	if e, ok := status.FromError(err); ok {
		assert.Equal(t, codes.PermissionDenied, e.Code())
	} else {
		t.Fatal("Error code was not retrieved")
	}
}
