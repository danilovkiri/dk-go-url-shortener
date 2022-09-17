package grpc

import (
	"context"
	"time"

	"github.com/danilovkiri/dk_go_url_shortener/internal/api/grpc/handlers"
	pb "github.com/danilovkiri/dk_go_url_shortener/internal/api/grpc/proto"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/shortener/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1"
)

var (
	serverStart = time.Now()
)

// uptime returns time in seconds since the server start-up.
func uptime() int64 {
	return int64(time.Since(serverStart).Seconds())
}

// ShortenerServer defines server methods and attributes.
type ShortenerServer struct {
	pb.UnimplementedShortenerServer
	grpcHandler *handlers.GRPCHandler
}

// InitServer returns a ShortenerServer object ready to be listening and serving.
func InitServer(ctx context.Context, cfg *config.Config, storage storage.URLStorage) (server *ShortenerServer, err error) {
	shortenerService, err := shortener.InitShortener(storage)
	if err != nil {
		return nil, err
	}
	grpcHandler, err := handlers.InitGRPCHandler(shortenerService, cfg)
	if err != nil {
		return nil, err
	}
	return &ShortenerServer{grpcHandler: grpcHandler}, nil
}

// GetUptime is a GRPC method for getting server uptime data.
func (s *ShortenerServer) GetUptime(_ context.Context, _ *pb.GetUptimeRequest) (*pb.GetUptimeResponse, error) {
	var response pb.GetUptimeResponse
	response.Uptime = uptime()
	return &response, nil
}

// PingDB is a GRPC method to check DB connection and establish it if closed.
func (s *ShortenerServer) PingDB(_ context.Context, _ *pb.PingDBRequest) (*pb.PingDBResponse, error) {
	result, err := s.grpcHandler.HandlePingDB()
	return result, err
}

// GetStats is a GRPC method to retrieve storage usage stats.
func (s *ShortenerServer) GetStats(ctx context.Context, _ *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	result, err := s.grpcHandler.HandleGetStats(ctx)
	return result, err
}

// GetURL is a GRPC method for getting original URL based on shortened URL ID.
func (s *ShortenerServer) GetURL(ctx context.Context, request *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	result, err := s.grpcHandler.HandleGetURL(ctx, request)
	return result, err
}

// GetURLsByUserID is a GRPC method for getting all user-specific pairs of full and shortened URLs.
func (s *ShortenerServer) GetURLsByUserID(ctx context.Context, _ *pb.GetURLsByUserIDRequest) (*pb.GetURLsByUserIDResponse, error) {
	result, err := s.grpcHandler.HandleGetURLsByUserID(ctx)
	return result, err
}

// PostURL is a GRPC method to get a shortened URL for an original URL and store them in DB.
func (s *ShortenerServer) PostURL(ctx context.Context, request *pb.PostURLRequest) (*pb.PostURLResponse, error) {
	result, err := s.grpcHandler.HandlePostURL(ctx, request)
	return result, err
}

// PostURLBatch is a GRPC method to get shortened URLs for a batch of original URLs and store them in DB.
func (s *ShortenerServer) PostURLBatch(ctx context.Context, request *pb.PostURLBatchRequest) (*pb.PostURLBatchResponse, error) {
	result, err := s.grpcHandler.HandlePostURLBatch(ctx, request)
	return result, err
}

// DeleteURLBatch is a GRPC method for deleting DB entries based on a batch of shortened URL IDs.
func (s *ShortenerServer) DeleteURLBatch(ctx context.Context, request *pb.DeleteURLBatchRequest) (*pb.DeleteURLBatchResponse, error) {
	result, err := s.grpcHandler.HandleDeleteURLBatch(ctx, request)
	return result, err
}
