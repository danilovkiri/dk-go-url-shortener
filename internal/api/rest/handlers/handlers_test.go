package handlers

import (
	"context"
	"encoding/json"
	"github.com/danilovkiri/dk_go_url_shortener/internal/api/rest/middleware"
	"github.com/danilovkiri/dk_go_url_shortener/internal/api/rest/modeldto"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/secretary/v1"
	shortenerService "github.com/danilovkiri/dk_go_url_shortener/internal/service/shortener"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/shortener/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/inpsql"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type HandlersTestSuite struct {
	suite.Suite
	storage          storage.URLStorage
	shortenerService shortenerService.Processor
	urlHandler       *URLHandler
	cookieHandler    *middleware.CookieHandler
	secretaryService *secretary.Secretary
	router           *chi.Mux
	ts               *httptest.Server
	ctx              context.Context
	cancel           context.CancelFunc
	wg               *sync.WaitGroup
}

func (suite *HandlersTestSuite) SetupTest() {
	cfg, _ := config.NewDefaultConfiguration()
	// necessary to set default parameters here since they are set in cfg.ParseFlags() which causes error
	cfg.ServerConfig.ServerAddress = ":8080"
	cfg.ServerConfig.BaseURL = "http://localhost:8080"
	cfg.StorageConfig.FileStoragePath = "url_storage.json"
	log.Println(cfg.StorageConfig.DatabaseDSN)
	if cfg.StorageConfig.DatabaseDSN == "" {
		cfg.StorageConfig.DatabaseDSN = "postgres://kirilldanilov:12345@localhost:5432/urlshort"
	}
	log.Println(cfg.StorageConfig.DatabaseDSN)
	suite.ctx, suite.cancel = context.WithCancel(context.Background())
	suite.wg = &sync.WaitGroup{}
	suite.wg.Add(1)
	suite.storage, _ = inpsql.InitStorage(suite.ctx, suite.wg, cfg.StorageConfig)
	suite.shortenerService, _ = shortener.InitShortener(suite.storage)
	suite.urlHandler, _ = InitURLHandler(suite.shortenerService, cfg.ServerConfig)
	suite.secretaryService, _ = secretary.NewSecretaryService(cfg.SecretConfig)
	suite.cookieHandler, _ = middleware.NewCookieHandler(suite.secretaryService, cfg.SecretConfig)
	suite.router = chi.NewRouter()
	suite.ts = httptest.NewServer(suite.router)
}

// TestHandlersTestSuite initializes test suite for being accessible
func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

func (suite *HandlersTestSuite) TestHandleGetURL() {
	userID := suite.secretaryService.Encode(uuid.New().String())
	sURL1, _ := suite.shortenerService.Encode(suite.ctx, "https://www.get-smth-1.com", userID)
	sURL2, _ := suite.shortenerService.Encode(suite.ctx, "https://www.get-smth-2.com", userID)
	suite.shortenerService.Delete(suite.ctx, []string{sURL2}, userID)
	// sleep here due to async deletion and wait for timer to trigger deletion
	time.Sleep(11 * time.Second)
	suite.router.Get("/{urlID}", suite.urlHandler.HandleGetURL())

	// set tests' parameters
	type want struct {
		code int
	}
	tests := []struct {
		name string
		sURL string
		want want
	}{
		{
			name: "Correct GET query",
			sURL: sURL1,
			want: want{
				code: 307,
			},
		},
		{
			name: "Invalid GET query",
			sURL: "",
			want: want{
				code: 404,
			},
		},
		{
			name: "Deleted GET query",
			sURL: sURL2,
			want: want{
				code: 410,
			},
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			client := resty.New()
			client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}))
			res, err := client.R().SetPathParams(map[string]string{"urlID": tt.sURL}).Get(suite.ts.URL + "/{urlID}")
			if err != nil {
				t.Fatalf(err.Error())
			}
			assert.Equal(t, tt.want.code, res.StatusCode())
		})
	}
	defer suite.ts.Close()
	suite.cancel()
	suite.wg.Wait()
}

func (suite *HandlersTestSuite) TestHandlePostURL() {
	suite.router.Use(suite.cookieHandler.CookieHandle)
	suite.router.Post("/", suite.urlHandler.HandlePostURL())

	// set tests' parameters
	type want struct {
		code int
	}
	tests := []struct {
		name string
		URL  string
		want want
	}{
		{
			name: "Correct POST query",
			URL:  "https://www.post-smth-1.com",
			want: want{
				code: 201,
			},
		},
		{
			name: "Invalid POST query (empty query)",
			URL:  "",
			want: want{
				code: 400,
			},
		},
		{
			name: "Invalid POST query (conflict)",
			URL:  "https://www.post-smth-1.com",
			want: want{
				code: 409,
			},
		},
		{
			name: "Invalid POST query (not URL)",
			URL:  "kke738enb734b",
			want: want{
				code: 400,
			},
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			payload := strings.NewReader(tt.URL)
			client := resty.New()
			res, err := client.R().SetBody(payload).Post(suite.ts.URL)
			if err != nil {
				t.Fatalf("Could not create POST request")
			}
			assert.Equal(t, tt.want.code, res.StatusCode())
		})
	}
	defer suite.ts.Close()
	suite.cancel()
	suite.wg.Wait()
}

func (suite *HandlersTestSuite) TestJSONHandlePostURL() {
	suite.router.Use(suite.cookieHandler.CookieHandle)
	suite.router.Post("/api/shorten", suite.urlHandler.JSONHandlePostURL())

	// set tests' parameters
	type want struct {
		code int
	}
	tests := []struct {
		name string
		URL  modeldto.RequestURL
		want want
	}{
		{
			name: "Correct POST query",
			URL: modeldto.RequestURL{
				URL: "https://www.post-smth-2.com",
			},
			want: want{
				code: 201,
			},
		},
		{
			name: "Invalid POST query (empty query)",
			URL: modeldto.RequestURL{
				URL: "",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "Invalid POST query (conflict)",
			URL: modeldto.RequestURL{
				URL: "https://www.post-smth-2.com",
			},
			want: want{
				code: 409,
			},
		},
		{
			name: "Invalid POST query (not URL)",
			URL: modeldto.RequestURL{
				URL: "kke738enb734b",
			},
			want: want{
				code: 400,
			},
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.URL)
			payload := strings.NewReader(string(reqBody))
			client := resty.New()
			res, err := client.R().SetBody(payload).Post(suite.ts.URL + "/api/shorten")
			if err != nil {
				t.Fatalf("Could not perform JSON POST request")
			}
			t.Logf(string(res.Body()))
			assert.Equal(t, tt.want.code, res.StatusCode())
		})
	}
	defer suite.ts.Close()
	suite.cancel()
	suite.wg.Wait()
}

func (suite *HandlersTestSuite) TestHandleGetURLsByUserID() {
	suite.router.Use(suite.cookieHandler.CookieHandle)
	userID1 := suite.secretaryService.Encode(uuid.New().String())
	userID2 := suite.secretaryService.Encode(uuid.New().String())
	_, _ = suite.shortenerService.Encode(suite.ctx, "https://www.get-smth-3.com", userID1)
	suite.router.Get("/api/user/urls", suite.urlHandler.HandleGetURLsByUserID())

	// set tests' parameters
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		token string
		want  want
	}{
		{
			name:  "Non-empty GET query",
			token: userID1,
			want: want{
				code: 200,
			},
		},
		{
			name:  "Empty GET query",
			token: userID2,
			want: want{
				code: 204,
			},
		},
		{
			name:  "Unauthorized GET query",
			token: "some_irrelevant_token",
			want: want{
				code: 401,
			},
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			client := resty.New()
			client.SetCookie(&http.Cookie{
				Name:  "user",
				Value: tt.token,
				Path:  "/",
			})
			res, err := client.R().Get(suite.ts.URL + "/api/user/urls")
			if err != nil {
				t.Fatalf("Could not perform GET by userID request")
			}
			assert.Equal(t, tt.want.code, res.StatusCode())
		})
	}
	defer suite.ts.Close()
	suite.cancel()
	suite.wg.Wait()
}

func (suite *HandlersTestSuite) TestJSONHandlePostURLBatch() {
	suite.router.Use(suite.cookieHandler.CookieHandle)
	suite.router.Post("/api/shorten/batch", suite.urlHandler.JSONHandlePostURLBatch())

	// set tests' parameters
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		batch []modeldto.RequestBatchURL
		want  want
	}{
		{
			name: "Correct POST batch query",
			batch: []modeldto.RequestBatchURL{
				{
					CorrelationID: "test1",
					URL:           "https://www.post-smth-3.com",
				},
				{
					CorrelationID: "test2",
					URL:           "https://www.post-smth-4.com",
				},
				{
					CorrelationID: "test3_conflict",
					URL:           "https://www.post-smth-2.com",
				},
			},
			want: want{
				code: 201,
			},
		},
		{
			name:  "Empty POST batch query",
			batch: []modeldto.RequestBatchURL{},
			want: want{
				code: 400,
			},
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.batch)
			payload := strings.NewReader(string(reqBody))
			client := resty.New()
			res, err := client.R().SetBody(payload).Post(suite.ts.URL + "/api/shorten/batch")
			if err != nil {
				t.Fatalf("Could not perform JSON POST request")
			}
			t.Logf(string(res.Body()))
			assert.Equal(t, tt.want.code, res.StatusCode())
		})
	}
	defer suite.ts.Close()
	suite.cancel()
	suite.wg.Wait()
}

func (suite *HandlersTestSuite) TestHandleDeleteURLBatch() {
	suite.router.Use(suite.cookieHandler.CookieHandle)
	suite.router.Delete("/api/user/urls", suite.urlHandler.HandleDeleteURLBatch())

	// set tests' parameters
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		batch []string
		want  want
	}{
		{
			name:  "Correct DELETE batch request",
			batch: []string{"hdsf6sd5f", "dsf6sd5f"},
			want: want{
				code: 202,
			},
		},
		{
			name:  "Empty DELETE batch request",
			batch: []string{},
			want: want{
				code: 202,
			},
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.batch)
			payload := strings.NewReader(string(reqBody))
			client := resty.New()
			res, err := client.R().SetBody(payload).Delete(suite.ts.URL + "/api/user/urls")
			if err != nil {
				t.Fatalf("Could not perform DELETE request")
			}
			t.Logf(string(res.Body()))
			assert.Equal(t, tt.want.code, res.StatusCode())
		})
	}
	defer suite.ts.Close()
	suite.cancel()
	suite.wg.Wait()
}

func (suite *HandlersTestSuite) TestHandlePingDB() {
	suite.router.Use(suite.cookieHandler.CookieHandle)
	suite.router.Get("/ping", suite.urlHandler.HandlePingDB())

	// set tests' parameters
	type want struct {
		code int
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "Ping 1",
			want: want{
				code: 200,
			},
		},
		{
			name: "Ping 2",
			want: want{
				code: 200,
			},
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			client := resty.New()
			res, err := client.R().Get(suite.ts.URL + "/ping")
			if err != nil {
				t.Fatalf("Could not perform DB ping")
			}
			t.Logf(string(res.Body()))
			assert.Equal(t, tt.want.code, res.StatusCode())
		})
	}
	defer suite.ts.Close()
	suite.cancel()
	suite.wg.Wait()
}
