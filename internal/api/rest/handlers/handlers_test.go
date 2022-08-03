package handlers

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/danilovkiri/dk_go_url_shortener/internal/api/rest/middleware"
	"github.com/danilovkiri/dk_go_url_shortener/internal/api/rest/modeldto"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/secretary/v1"
	shortenerService "github.com/danilovkiri/dk_go_url_shortener/internal/service/shortener"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/shortener/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/infile"
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
	// parsing flags causes flag redefined errors
	//cfg.ParseFlags()
	suite.ctx, suite.cancel = context.WithCancel(context.Background())
	suite.wg = &sync.WaitGroup{}
	suite.wg.Add(1)
	suite.storage, _ = infile.InitStorage(suite.ctx, suite.wg, cfg.StorageConfig)
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
	sURL, _ := suite.shortenerService.Encode(suite.ctx, "https://www.yandex.ru", userID)
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
			sURL: sURL,
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
			URL:  "https://www.yandex.az",
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
				URL: "https://www.yandex.kz",
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
	userIDFull := suite.secretaryService.Encode(uuid.New().String())
	userIDEmpty := suite.secretaryService.Encode(uuid.New().String())
	_, _ = suite.shortenerService.Encode(suite.ctx, "https://www.yandex.nd", userIDFull)
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
			token: userIDFull,
			want: want{
				code: 200,
			},
		},
		{
			name:  "Empty GET query",
			token: userIDEmpty,
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
					URL:           "https://www.kinopoisk.ru",
				},
				{
					CorrelationID: "test2",
					URL:           "https://www.vk.com",
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
	client := resty.New()
	// perform each test
	for i := 0; i < 10; i++ {
		suite.T().Run("ping", func(t *testing.T) {
			res, err := client.R().Get(suite.ts.URL + "/ping")
			if err != nil {
				t.Fatalf("Could not perform GET request")
			}
			assert.Equal(t, 200, res.StatusCode())
		})
	}
	defer suite.ts.Close()
	suite.cancel()
	suite.wg.Wait()
}

func BenchmarkInitURLHandler(b *testing.B) {
	cfg, _ := config.NewDefaultConfiguration()
	cfg.ServerConfig.ServerAddress = ":8080"
	cfg.ServerConfig.BaseURL = "http://localhost:8080"
	cfg.StorageConfig.FileStoragePath = "url_storage.json"
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg.StorageConfig)
	svc, _ := shortener.InitShortener(strg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = InitURLHandler(svc, cfg.ServerConfig)
	}
}

func BenchmarkURLHandler_HandleGetURL(b *testing.B) {
	cfg, _ := config.NewDefaultConfiguration()
	cfg.ServerConfig.ServerAddress = ":8080"
	cfg.ServerConfig.BaseURL = "http://localhost:8080"
	cfg.StorageConfig.FileStoragePath = "url_storage.json"
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg.StorageConfig)
	svc, _ := shortener.InitShortener(strg)
	urlHandler, _ := InitURLHandler(svc, cfg.ServerConfig)
	secretaryService, _ := secretary.NewSecretaryService(cfg.SecretConfig)
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	userID := secretaryService.Encode(uuid.New().String())
	sURL, _ := svc.Encode(ctx, "https://www.yandex.ru", userID)
	router.Get("/{urlID}", urlHandler.HandleGetURL())
	client := resty.New()
	client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.R().SetPathParams(map[string]string{"urlID": sURL}).Get(ts.URL + "/{urlID}")
	}
}

func BenchmarkURLHandler_HandleGetURLsByUserID(b *testing.B) {
	cfg, _ := config.NewDefaultConfiguration()
	cfg.ServerConfig.ServerAddress = ":8080"
	cfg.ServerConfig.BaseURL = "http://localhost:8080"
	cfg.StorageConfig.FileStoragePath = "url_storage.json"
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg.StorageConfig)
	svc, _ := shortener.InitShortener(strg)
	urlHandler, _ := InitURLHandler(svc, cfg.ServerConfig)
	secretaryService, _ := secretary.NewSecretaryService(cfg.SecretConfig)
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg.SecretConfig)
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	router.Use(cookieHandler.CookieHandle)
	userIDFull := secretaryService.Encode(uuid.New().String())
	_, _ = svc.Encode(ctx, "https://www.yandex.nd", userIDFull)
	router.Get("/api/user/urls", urlHandler.HandleGetURLsByUserID())
	client := resty.New()
	client.SetCookie(&http.Cookie{
		Name:  "user",
		Value: userIDFull,
		Path:  "/",
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.R().Get(ts.URL + "/api/user/urls")
	}
}

func BenchmarkURLHandler_HandlePostURL(b *testing.B) {
	cfg, _ := config.NewDefaultConfiguration()
	cfg.ServerConfig.ServerAddress = ":8080"
	cfg.ServerConfig.BaseURL = "http://localhost:8080"
	cfg.StorageConfig.FileStoragePath = "url_storage.json"
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg.StorageConfig)
	svc, _ := shortener.InitShortener(strg)
	urlHandler, _ := InitURLHandler(svc, cfg.ServerConfig)
	secretaryService, _ := secretary.NewSecretaryService(cfg.SecretConfig)
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg.SecretConfig)
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	router.Use(cookieHandler.CookieHandle)
	router.Post("/", urlHandler.HandlePostURL())
	payload := strings.NewReader("https://www.some-url.com")
	client := resty.New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.R().SetBody(payload).Post(ts.URL)
	}
}

func BenchmarkURLHandler_JSONHandlePostURL(b *testing.B) {
	cfg, _ := config.NewDefaultConfiguration()
	cfg.ServerConfig.ServerAddress = ":8080"
	cfg.ServerConfig.BaseURL = "http://localhost:8080"
	cfg.StorageConfig.FileStoragePath = "url_storage.json"
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg.StorageConfig)
	svc, _ := shortener.InitShortener(strg)
	urlHandler, _ := InitURLHandler(svc, cfg.ServerConfig)
	secretaryService, _ := secretary.NewSecretaryService(cfg.SecretConfig)
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg.SecretConfig)
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	router.Use(cookieHandler.CookieHandle)
	router.Post("/api/shorten", urlHandler.JSONHandlePostURL())
	client := resty.New()
	client.SetCookieJar(nil)
	b.ResetTimer()
	b.Run("subtask", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			URL := modeldto.RequestURL{
				URL: "https://www." + randStringBytes(10) + ".com",
			}
			reqBody, _ := json.Marshal(URL)
			payload := strings.NewReader(string(reqBody))
			b.StartTimer()
			_, _ = client.R().SetBody(payload).Post(ts.URL + "/api/shorten")
		}
	})

	cancel()
	wg.Wait()
}

func BenchmarkURLHandler_HandlePingDB(b *testing.B) {
	cfg, _ := config.NewDefaultConfiguration()
	cfg.ServerConfig.ServerAddress = ":8080"
	cfg.ServerConfig.BaseURL = "http://localhost:8080"
	cfg.StorageConfig.FileStoragePath = "url_storage.json"
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg.StorageConfig)
	svc, _ := shortener.InitShortener(strg)
	urlHandler, _ := InitURLHandler(svc, cfg.ServerConfig)
	secretaryService, _ := secretary.NewSecretaryService(cfg.SecretConfig)
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg.SecretConfig)
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	router.Use(cookieHandler.CookieHandle)
	router.Post("/ping", urlHandler.HandlePingDB())
	client := resty.New()
	client.SetCookieJar(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.R().Get(ts.URL + "/ping")
	}
	cancel()
	wg.Wait()
}

func BenchmarkURLHandler_JSONHandlePostURLBatch(b *testing.B) {
	cfg, _ := config.NewDefaultConfiguration()
	cfg.ServerConfig.ServerAddress = ":8080"
	cfg.ServerConfig.BaseURL = "http://localhost:8080"
	cfg.StorageConfig.FileStoragePath = "url_storage.json"
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg.StorageConfig)
	svc, _ := shortener.InitShortener(strg)
	urlHandler, _ := InitURLHandler(svc, cfg.ServerConfig)
	secretaryService, _ := secretary.NewSecretaryService(cfg.SecretConfig)
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg.SecretConfig)
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	router.Use(cookieHandler.CookieHandle)
	router.Post("/api/shorten/batch", urlHandler.JSONHandlePostURLBatch())
	client := resty.New()
	client.SetCookieJar(nil)
	b.ResetTimer()
	b.Run("subtask", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			batch := []modeldto.RequestBatchURL{
				{
					CorrelationID: "test1",
					URL:           "https://www." + randStringBytes(10) + ".com",
				},
				{
					CorrelationID: "test2",
					URL:           "https://www." + randStringBytes(10) + ".com",
				},
			}
			reqBody, _ := json.Marshal(batch)
			payload := strings.NewReader(string(reqBody))
			b.StartTimer()
			_, _ = client.R().SetBody(payload).Post(ts.URL + "/api/shorten/batch")
		}
	})
	cancel()
	wg.Wait()
}

func BenchmarkURLHandler_HandleDeleteURLBatch(b *testing.B) {
	cfg, _ := config.NewDefaultConfiguration()
	cfg.ServerConfig.ServerAddress = ":8080"
	cfg.ServerConfig.BaseURL = "http://localhost:8080"
	cfg.StorageConfig.FileStoragePath = "url_storage.json"
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg.StorageConfig)
	svc, _ := shortener.InitShortener(strg)
	urlHandler, _ := InitURLHandler(svc, cfg.ServerConfig)
	secretaryService, _ := secretary.NewSecretaryService(cfg.SecretConfig)
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg.SecretConfig)
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	router.Use(cookieHandler.CookieHandle)
	router.Delete("/api/user/urls", urlHandler.JSONHandlePostURLBatch())
	client := resty.New()
	client.SetCookieJar(nil)
	b.ResetTimer()
	b.Run("subtask", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			batch := []string{randStringBytes(10), randStringBytes(10), randStringBytes(10)}
			reqBody, _ := json.Marshal(batch)
			payload := strings.NewReader(string(reqBody))
			b.StartTimer()
			_, _ = client.R().SetBody(payload).Post(ts.URL + "/api/shorten/batch")
		}
	})
	cancel()
	wg.Wait()
}

func randStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
