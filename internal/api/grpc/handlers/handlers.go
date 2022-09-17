package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/danilovkiri/dk_go_url_shortener/internal/api/grpc/interceptors"
	pb "github.com/danilovkiri/dk_go_url_shortener/internal/api/grpc/proto"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/shortener"
	storageErrors "github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCHandler defines data structure handling and provides support for adding new implementations.
type GRPCHandler struct {
	processor    shortener.Processor
	serverConfig *config.Config
}

// InitGRPCHandler initializes a URLHandler object and sets its attributes.
func InitGRPCHandler(processor shortener.Processor, serverConfig *config.Config) (*GRPCHandler, error) {
	if processor == nil {
		return nil, fmt.Errorf("nil Shortener Service was passed to service URL Handler initializer")
	}
	return &GRPCHandler{processor: processor, serverConfig: serverConfig}, nil
}

// HandlePingDB handles PSQL DB pinging to check connection status.
func (h *GRPCHandler) HandlePingDB() (*pb.PingDBResponse, error) {
	err := h.processor.PingDB()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var response pb.PingDBResponse
	return &response, nil
}

// HandleGetStats provides client with statistics on URLs and clients.
func (h *GRPCHandler) HandleGetStats(ctx context.Context) (*pb.GetStatsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	nURLs, nUsers, err := h.processor.GetStats(ctx)
	if err != nil {
		var contextTimeoutExceededError *storageErrors.ContextTimeoutExceededError
		if errors.As(err, &contextTimeoutExceededError) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	response := pb.GetStatsResponse{
		Users: nUsers,
		Urls:  nURLs,
	}
	return &response, nil
}

// HandleGetURL provides client with a redirect to the original URL accessed by shortened URL.
func (h *GRPCHandler) HandleGetURL(ctx context.Context, request *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	sURL := request.ShortUrlId
	URL, err := h.processor.Decode(ctx, sURL)
	if err != nil {
		var contextTimeoutExceededError *storageErrors.ContextTimeoutExceededError
		var deletedError *storageErrors.DeletedError
		if errors.As(err, &contextTimeoutExceededError) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		} else if errors.As(err, &deletedError) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	response := pb.GetURLResponse{
		RedirectTo: URL,
	}
	return &response, nil
}

// HandleGetURLsByUserID provides shortening service using modeldto.ResponseFullURL schema.
func (h *GRPCHandler) HandleGetURLsByUserID(ctx context.Context) (*pb.GetURLsByUserIDResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	URLs, err := h.processor.DecodeByUserID(ctx, userID)
	if err != nil {
		var contextTimeoutExceededError *storageErrors.ContextTimeoutExceededError
		if errors.As(err, &contextTimeoutExceededError) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(URLs) == 0 {
		return nil, status.Error(codes.NotFound, `No content available`)
	}
	u, err := url.Parse(h.serverConfig.BaseURL)
	if err != nil {
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

// HandlePostURL stores the original URL with its shortened version.
func (h *GRPCHandler) HandlePostURL(ctx context.Context, request *pb.PostURLRequest) (*pb.PostURLResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	URL := request.FullUrl
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	u, err := url.Parse(h.serverConfig.BaseURL)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	sURL, err := h.processor.Encode(ctx, URL, userID)
	if err != nil {
		var contextTimeoutExceededError *storageErrors.ContextTimeoutExceededError
		var alreadyExistsError *storageErrors.AlreadyExistsError
		if errors.As(err, &contextTimeoutExceededError) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		} else if errors.As(err, &alreadyExistsError) {
			u.Path = alreadyExistsError.ValidSURL
			response := pb.PostURLResponse{
				ShortUrl: u.String(),
			}
			return &response, status.Error(codes.AlreadyExists, `Entry already exists and was returned in response body`)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	u.Path = sURL
	response := pb.PostURLResponse{
		ShortUrl: u.String(),
	}
	return &response, nil
}

// HandlePostURLBatch provides shortening service for batch processing using PostURLBatchRequest and PostURLBatchResponse schemas.
func (h *GRPCHandler) HandlePostURLBatch(ctx context.Context, request *pb.PostURLBatchRequest) (*pb.PostURLBatchResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	u, err := url.Parse(h.serverConfig.BaseURL)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response := pb.PostURLBatchResponse{}
	for _, requestBatchURL := range request.RequestUrls {
		sURL, err1 := h.processor.Encode(ctx, requestBatchURL.Url, userID)
		if err1 != nil {
			var contextTimeoutExceededError *storageErrors.ContextTimeoutExceededError
			var alreadyExistsError *storageErrors.AlreadyExistsError
			if errors.As(err1, &contextTimeoutExceededError) {
				return nil, status.Error(codes.DeadlineExceeded, err.Error())
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
			return nil, status.Error(codes.Internal, err.Error())
		}
		u.Path = sURL
		responseBatchURL := pb.PostURLBatch{
			CorrelationId: requestBatchURL.CorrelationId,
			Url:           u.String(),
		}
		response.ResponseUrls = append(response.ResponseUrls, &responseBatchURL)
	}
	return &response, nil
}

// HandleDeleteURLBatch sets a tag for deletion for a batch of URL entries in DB.
func (h *GRPCHandler) HandleDeleteURLBatch(ctx context.Context, request *pb.DeleteURLBatchRequest) (*pb.DeleteURLBatchResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	deleteURLs := make([]string, 0)
	for _, requestURL := range request.RequestUrls.Urls {
		deleteURLs = append(deleteURLs, requestURL)
	}
	h.processor.Delete(ctx, deleteURLs, userID)
	var response pb.DeleteURLBatchResponse
	return &response, nil
}

// getUserID retrieves user identifier as a value of GRPC metadata.
func getUserID(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("GRPC metadata was not found")
	}
	values := md.Get(interceptors.UserAuthKey)
	if len(values) <= 0 {
		return "", errors.New("empty array of values was found for user key")
	}
	userID := values[0]
	return userID, nil
}
