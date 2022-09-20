package handlers

import (
	"context"
	"log"
	"net"
	"sync"
	"testing"

	"github.com/danilovkiri/dk_go_url_shortener/internal/api/grpc/interceptors"
	pb "github.com/danilovkiri/dk_go_url_shortener/internal/api/grpc/proto"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/secretary/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/infile"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type HandlersTestSuite struct {
	suite.Suite
	storage          storage.URLStorage
	authHandler      *interceptors.AuthHandler
	secretaryService *secretary.Secretary
	router           *chi.Mux
	ctx              context.Context
	cancel           context.CancelFunc
	wg               *sync.WaitGroup
	server           *ShortenerServer
	s                *grpc.Server
}

func (suite *HandlersTestSuite) SetupTest() {
	cfg := config.NewDefaultConfiguration()
	// necessary to set default parameters here since they are set in cfg.ParseFlags() which causes error
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	cfg.UserKey = "jds__63h3_7ds"
	cfg.AuthKey = "user"
	// parsing flags causes flag redefined errors
	//cfg.ParseFlags()
	suite.ctx, suite.cancel = context.WithCancel(context.Background())
	suite.wg = &sync.WaitGroup{}
	suite.wg.Add(1)
	suite.storage, _ = infile.InitStorage(suite.ctx, suite.wg, cfg)
	suite.server, _ = InitServer(suite.ctx, cfg, suite.storage)
	suite.secretaryService = secretary.NewSecretaryService(cfg)
	suite.authHandler = interceptors.NewAuthHandler(suite.secretaryService, cfg)
	suite.router = chi.NewRouter()
	suite.s = grpc.NewServer(grpc.UnaryInterceptor(suite.authHandler.UnaryServerInterceptor()))
	pb.RegisterShortenerServer(suite.s, suite.server)
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	go suite.s.Serve(listen)
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

func (suite *HandlersTestSuite) TestPingDB() {
	// create a client
	conn, err := grpc.Dial(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	token := "8773a90a68ebd0fd56dffb1441682414fbec5f454eba9be6129bb00744f50d7f19fd870e97eba101a03b857c675e4836de6f5196"
	md := metadata.New(map[string]string{"user": token})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	c := pb.NewShortenerClient(conn)

	// perform each test
	for i := 0; i < 1; i++ {
		suite.T().Run("ping", func(t *testing.T) {
			var request emptypb.Empty
			resp, err1 := c.PingDB(ctx, &request)
			if err1 != nil {
				t.Fatalf("Could not perform request: %s", err1)
			}
			assert.Equal(t, nil, err1)
			var response *emptypb.Empty
			assert.IsType(t, response, resp)
		})
	}
	suite.s.GracefulStop()
	suite.cancel()
	suite.wg.Wait()
}

func (suite *HandlersTestSuite) TestGetUptime() {
	// create a client
	conn, err := grpc.Dial(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	token := "8773a90a68ebd0fd56dffb1441682414fbec5f454eba9be6129bb00744f50d7f19fd870e97eba101a03b857c675e4836de6f5196"
	md := metadata.New(map[string]string{"user": token})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	c := pb.NewShortenerClient(conn)

	// perform each test
	for i := 0; i < 10; i++ {
		suite.T().Run("uptime", func(t *testing.T) {
			var request emptypb.Empty
			resp, err1 := c.GetUptime(ctx, &request)
			if err1 != nil {
				t.Fatalf("Could not perform request: %s", err1)
			}
			assert.Equal(t, nil, err1)
			assert.IsType(t, &pb.GetUptimeResponse{}, resp)
		})
	}
	suite.s.GracefulStop()
	suite.cancel()
	suite.wg.Wait()
}

func (suite *HandlersTestSuite) TestGetStats() {
	// create a client
	conn, err := grpc.Dial(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	token := "8773a90a68ebd0fd56dffb1441682414fbec5f454eba9be6129bb00744f50d7f19fd870e97eba101a03b857c675e4836de6f5196"
	md := metadata.New(map[string]string{"user": token})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	c := pb.NewShortenerClient(conn)

	// perform each test
	for i := 0; i < 10; i++ {
		suite.T().Run("stats", func(t *testing.T) {
			var request emptypb.Empty
			resp, err1 := c.GetStats(ctx, &request)
			if err1 != nil {
				t.Fatalf("Could not perform request: %s", err1)
			}
			assert.Equal(t, nil, err1)
			assert.IsType(t, &pb.GetStatsResponse{}, resp)
			assert.Equal(t, int64(0), resp.Urls)
			assert.Equal(t, int64(0), resp.Users)
		})
	}
	suite.s.GracefulStop()
	suite.cancel()
	suite.wg.Wait()
}

func (suite *HandlersTestSuite) TestPostURL() {
	// create a client
	conn, err := grpc.Dial(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	token := "8773a90a68ebd0fd56dffb1441682414fbec5f454eba9be6129bb00744f50d7f19fd870e97eba101a03b857c675e4836de6f5196"
	md := metadata.New(map[string]string{"user": token})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	c := pb.NewShortenerClient(conn)

	// set tests' parameters
	type want struct {
		code codes.Code
	}
	tests := []struct {
		name string
		URL  string
		want want
	}{
		{
			name: "Correct POST query",
			URL:  "https://www.yandex.ru",
			want: want{
				code: codes.OK,
			},
		},
		{
			name: "Invalid POST query (empty query)",
			URL:  "",
			want: want{
				code: codes.Internal,
			},
		},
		{
			name: "Invalid POST query (not URL)",
			URL:  "kke738enb734b",
			want: want{
				code: codes.Internal,
			},
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run("post", func(t *testing.T) {
			resp, err1 := c.PostURL(ctx, &pb.PostURLRequest{FullUrl: tt.URL})
			e, _ := status.FromError(err1)
			assert.Equal(t, tt.want.code, e.Code())
			assert.IsType(t, &pb.PostURLResponse{}, resp)
		})
	}
	suite.s.GracefulStop()
	suite.cancel()
	suite.wg.Wait()
}

func (suite *HandlersTestSuite) TestPostURLBatch() {
	// create a client
	conn, err := grpc.Dial(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	token := "8773a90a68ebd0fd56dffb1441682414fbec5f454eba9be6129bb00744f50d7f19fd870e97eba101a03b857c675e4836de6f5196"
	md := metadata.New(map[string]string{"user": token})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	c := pb.NewShortenerClient(conn)

	// set tests' parameters
	type want struct {
		code codes.Code
	}
	tests := []struct {
		name  string
		batch []*pb.PostURLBatch
		want  want
	}{
		{
			name: "Correct POST batch query",
			batch: []*pb.PostURLBatch{
				{
					CorrelationId: "test1",
					Url:           "https://www.kinopoisk.ru",
				},
				{
					CorrelationId: "test2",
					Url:           "https://www.vk.com",
				},
			},
			want: want{
				code: codes.OK,
			},
		},
		{
			name:  "Empty POST batch query",
			batch: []*pb.PostURLBatch{},
			want: want{
				code: codes.Internal,
			},
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run("post", func(t *testing.T) {
			resp, err1 := c.PostURLBatch(ctx, &pb.PostURLBatchRequest{RequestUrls: tt.batch})
			e, _ := status.FromError(err1)
			assert.Equal(t, tt.want.code, e.Code())
			assert.IsType(t, &pb.PostURLBatchResponse{}, resp)
		})
	}
	suite.s.GracefulStop()
	suite.cancel()
	suite.wg.Wait()
}

func (suite *HandlersTestSuite) TestGetURL() {
	// create a client
	conn, err := grpc.Dial(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	token := "8773a90a68ebd0fd56dffb1441682414fbec5f454eba9be6129bb00744f50d7f19fd870e97eba101a03b857c675e4836de6f5196"
	md := metadata.New(map[string]string{"user": token})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	c := pb.NewShortenerClient(conn)

	// set tests' parameters
	sURL, _ := suite.server.processor.Encode(suite.ctx, "https://www.yandex.nd", token)
	type want struct {
		code codes.Code
	}
	tests := []struct {
		name string
		sURL string
		want want
	}{
		{
			name: "Correct GET query",
			sURL: sURL,
			want: want{
				code: codes.OK,
			},
		},
		{
			name: "Invalid GET query",
			sURL: "",
			want: want{
				code: codes.Internal,
			},
		},
		{
			name: "Absent GET query",
			sURL: "some_absent_sURL",
			want: want{
				code: codes.Internal,
			},
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run("post", func(t *testing.T) {
			resp, err1 := c.GetURL(ctx, &pb.GetURLRequest{ShortUrlId: tt.sURL})
			e, _ := status.FromError(err1)
			assert.Equal(t, tt.want.code, e.Code())
			assert.IsType(t, &pb.GetURLResponse{}, resp)
		})
	}
	suite.s.GracefulStop()
	suite.cancel()
	suite.wg.Wait()
}

func (suite *HandlersTestSuite) TestGetURLsByUserID() {
	// create a client
	conn, err := grpc.Dial(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	token1 := suite.secretaryService.Encode(uuid.New().String())
	token2 := suite.secretaryService.Encode(uuid.New().String())
	c := pb.NewShortenerClient(conn)

	// set tests' parameters
	_, _ = suite.server.processor.Encode(suite.ctx, "https://www.yandex.nd", token1)
	_, _ = suite.server.processor.Encode(suite.ctx, "https://www.yandex.kz", token1)
	_, _ = suite.server.processor.Encode(suite.ctx, "https://www.yandex.am", token1)
	type want struct {
		code codes.Code
	}
	tests := []struct {
		name  string
		token string
		want  want
	}{
		{
			name:  "Non-empty GET query",
			token: token1,
			want: want{
				code: codes.OK,
			},
		},
		{
			name:  "Empty GET query",
			token: token2,
			want: want{
				code: codes.NotFound,
			},
		},
		{
			name:  "Unauthorized GET query",
			token: "some_irrelevant_token",
			want: want{
				code: codes.PermissionDenied,
			},
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run("post", func(t *testing.T) {
			md := metadata.New(map[string]string{"user": tt.token})
			ctx := metadata.NewOutgoingContext(context.Background(), md)
			var request emptypb.Empty
			resp, err1 := c.GetURLsByUserID(ctx, &request)
			e, _ := status.FromError(err1)
			assert.Equal(t, tt.want.code, e.Code())
			assert.IsType(t, &pb.GetURLsByUserIDResponse{}, resp)
		})
	}
	suite.s.GracefulStop()
	suite.cancel()
	suite.wg.Wait()
}

func (suite *HandlersTestSuite) TestDeleteURLBatch() {
	// create a client
	conn, err := grpc.Dial(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	token := "8773a90a68ebd0fd56dffb1441682414fbec5f454eba9be6129bb00744f50d7f19fd870e97eba101a03b857c675e4836de6f5196"
	md := metadata.New(map[string]string{"user": token})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	c := pb.NewShortenerClient(conn)

	// set tests' parameters
	type want struct {
		code codes.Code
	}
	tests := []struct {
		name  string
		batch *pb.DeleteURLBatch
		want  want
	}{
		{
			name:  "Correct DELETE batch request",
			batch: &pb.DeleteURLBatch{Urls: []string{"hdsf6sd5f", "dsf6sd5f"}},
			want: want{
				code: codes.OK,
			},
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run("post", func(t *testing.T) {
			resp, err1 := c.DeleteURLBatch(ctx, &pb.DeleteURLBatchRequest{RequestUrls: tt.batch})
			e, _ := status.FromError(err1)
			assert.Equal(t, tt.want.code, e.Code())
			var response *emptypb.Empty
			assert.IsType(t, response, resp)
		})
	}
	suite.s.GracefulStop()
	suite.cancel()
	suite.wg.Wait()
}
