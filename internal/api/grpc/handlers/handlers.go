// Package handlers provides GRPC methods.
package handlers

import (
	"context"
	"errors"
	"log"
	"net/url"
	"time"

	pb "github.com/danilovkiri/dk_go_url_shortener/internal/api/grpc/proto"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	processor "github.com/danilovkiri/dk_go_url_shortener/internal/service/shortener"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/shortener/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1"
	storageErrors "github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
	processor processor.Processor
	cfg       *config.Config
}

// InitServer returns a ShortenerServer object ready to be listening and serving.
func InitServer(ctx context.Context, cfg *config.Config, storage storage.URLStorage) (server *ShortenerServer, err error) {
	shortenerService, err := shortener.InitShortener(storage)
	if err != nil {
		return nil, err
	}
	return &ShortenerServer{processor: shortenerService, cfg: cfg}, nil
}

// GetUptime is a GRPC method for getting server uptime data.
func (s *ShortenerServer) GetUptime(_ context.Context, _ *emptypb.Empty) (*pb.GetUptimeResponse, error) {
	var response pb.GetUptimeResponse
	response.Uptime = uptime()
	return &response, nil
}

// PingDB is a GRPC method to check DB connection and establish it if closed.
func (s *ShortenerServer) PingDB(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	err := s.processor.PingDB()
	if err != nil {
		log.Println("HandlePingDB:", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	var response emptypb.Empty
	return &response, nil
}

// GetStats is a GRPC method to retrieve storage usage stats.
func (s *ShortenerServer) GetStats(ctx context.Context, _ *emptypb.Empty) (*pb.GetStatsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	nURLs, nUsers, err := s.processor.GetStats(ctx)
	if err != nil {
		var contextTimeoutExceededError *storageErrors.ContextTimeoutExceededError
		if errors.As(err, &contextTimeoutExceededError) {
			log.Println("HandleGetStats:", err)
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}
		log.Println("HandleGetStats:", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	response := pb.GetStatsResponse{
		Users: nUsers,
		Urls:  nURLs,
	}
	return &response, nil
}

// GetURL is a GRPC method for getting original URL based on shortened URL ID.
func (s *ShortenerServer) GetURL(ctx context.Context, request *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	sURL := request.ShortUrlId
	URL, err := s.processor.Decode(ctx, sURL)
	if err != nil {
		var contextTimeoutExceededError *storageErrors.ContextTimeoutExceededError
		var deletedError *storageErrors.DeletedError
		if errors.As(err, &contextTimeoutExceededError) {
			log.Println("HandleGetURL:", err)
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		} else if errors.As(err, &deletedError) {
			log.Println("HandleGetURL:", err)
			return nil, status.Error(codes.NotFound, err.Error())
		}
		log.Println("HandleGetURL:", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	log.Println("HandleGetURL: retrieved URL", URL)
	response := pb.GetURLResponse{
		RedirectTo: URL,
	}
	return &response, nil
}

// GetURLsByUserID is a GRPC method for getting all user-specific pairs of full and shortened URLs.
func (s *ShortenerServer) GetURLsByUserID(ctx context.Context, _ *emptypb.Empty) (*pb.GetURLsByUserIDResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	userID := s.getUserID(ctx)
	URLs, err := s.processor.DecodeByUserID(ctx, userID)
	if err != nil {
		var contextTimeoutExceededError *storageErrors.ContextTimeoutExceededError
		if errors.As(err, &contextTimeoutExceededError) {
			log.Println("HandleGetURLsByUserID:", err)
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}
		log.Println("HandleGetURLsByUserID:", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(URLs) == 0 {
		log.Println("HandleGetURLsByUserID:", "No content available")
		return nil, status.Error(codes.NotFound, `No content available`)
	}
	u, err := url.Parse(s.cfg.BaseURL)
	if err != nil {
		log.Println("HandleGetURLsByUserID:", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	response := pb.GetURLsByUserIDResponse{}
	for _, fullURL := range URLs {
		u.Path = fullURL.SURL
		responseURL := pb.ResponsePairURL{
			FullUrl:  fullURL.URL,
			ShortUrl: u.String(),
		}
		response.ResponsePairsUrls = append(response.ResponsePairsUrls, &responseURL)
	}
	return &response, nil
}

// PostURL is a GRPC method to get a shortened URL for an original URL and store them in DB.
func (s *ShortenerServer) PostURL(ctx context.Context, request *pb.PostURLRequest) (*pb.PostURLResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	URL := request.FullUrl
	userID := s.getUserID(ctx)
	u, err := url.Parse(s.cfg.BaseURL)
	if err != nil {
		log.Println("HandlePostURL:", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	sURL, err := s.processor.Encode(ctx, URL, userID)
	if err != nil {
		var contextTimeoutExceededError *storageErrors.ContextTimeoutExceededError
		var alreadyExistsError *storageErrors.AlreadyExistsError
		if errors.As(err, &contextTimeoutExceededError) {
			log.Println("HandlePostURL:", err)
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		} else if errors.As(err, &alreadyExistsError) {
			u.Path = alreadyExistsError.ValidSURL
			response := pb.PostURLResponse{
				ShortUrl: u.String(),
			}
			return &response, status.Error(codes.AlreadyExists, `Entry already exists and was returned in response body`)
		}
		log.Println("HandlePostURL:", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	log.Println("HandlePostURL: stored", URL, "as", sURL)
	u.Path = sURL
	response := pb.PostURLResponse{
		ShortUrl: u.String(),
	}
	return &response, nil
}

// PostURLBatch is a GRPC method to get shortened URLs for a batch of original URLs and store them in DB.
func (s *ShortenerServer) PostURLBatch(ctx context.Context, request *pb.PostURLBatchRequest) (*pb.PostURLBatchResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	userID := s.getUserID(ctx)
	u, err := url.Parse(s.cfg.BaseURL)
	if err != nil {
		log.Println("HandlePostURLBatch:", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(request.RequestUrls) == 0 {
		log.Println("HandlePostURLBatch:", "empty request")
		return nil, status.Error(codes.Internal, "Empty request")
	}
	response := pb.PostURLBatchResponse{}
	for _, requestBatchURL := range request.RequestUrls {
		sURL, err1 := s.processor.Encode(ctx, requestBatchURL.Url, userID)
		if err1 != nil {
			var contextTimeoutExceededError *storageErrors.ContextTimeoutExceededError
			var alreadyExistsError *storageErrors.AlreadyExistsError
			if errors.As(err1, &contextTimeoutExceededError) {
				log.Println("HandlePostURLBatch:", err1)
				return nil, status.Error(codes.DeadlineExceeded, err1.Error())
			} else if errors.As(err1, &alreadyExistsError) {
				sURL = alreadyExistsError.ValidSURL
				u.Path = sURL
				responseBatchURL := pb.PostURLBatch{
					CorrelationId: requestBatchURL.CorrelationId,
					Url:           u.String(),
				}
				response.ResponseUrls = append(response.ResponseUrls, &responseBatchURL)
				continue
			}
			log.Println("HandlePostURLBatch:", err1)
			return nil, status.Error(codes.Internal, err1.Error())
		}
		log.Println("HandlePostURLBatch: stored", requestBatchURL.Url, "as", sURL)
		u.Path = sURL
		responseBatchURL := pb.PostURLBatch{
			CorrelationId: requestBatchURL.CorrelationId,
			Url:           u.String(),
		}
		response.ResponseUrls = append(response.ResponseUrls, &responseBatchURL)
	}
	return &response, nil
}

// DeleteURLBatch is a GRPC method for deleting DB entries based on a batch of shortened URL IDs.
func (s *ShortenerServer) DeleteURLBatch(ctx context.Context, request *pb.DeleteURLBatchRequest) (*emptypb.Empty, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	userID := s.getUserID(ctx)
	deleteURLs := make([]string, 0)
	deleteURLs = append(deleteURLs, request.RequestUrls.Urls...)
	log.Println("DELETE request detected for", deleteURLs)
	s.processor.Delete(ctx, deleteURLs, userID)
	var response emptypb.Empty
	return &response, nil
}

// getUserID retrieves user identifier as a value of GRPC metadata.
func (s *ShortenerServer) getUserID(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)
	values := md.Get(s.cfg.AuthKey)
	userID := values[0]
	return userID
}
