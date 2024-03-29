package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/danilovkiri/dk_go_url_shortener/internal/api/rest/middleware"
	"github.com/danilovkiri/dk_go_url_shortener/internal/api/rest/modeldto"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/secretary/v1"
	shortenerService "github.com/danilovkiri/dk_go_url_shortener/internal/service/shortener"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/shortener/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/infile"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func randStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// Tests

func TestInitURLHandler_Fail(t *testing.T) {
	cfg := config.NewDefaultConfiguration()
	// necessary to set default parameters here since they are set in cfg.ParseFlags() which causes error
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	_, err := InitURLHandler(nil, cfg)
	assert.Equal(t, fmt.Errorf("nil Shortener Service was passed to service URL Handler initializer"), err)
}

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
	cfg := config.NewDefaultConfiguration()
	// necessary to set default parameters here since they are set in cfg.ParseFlags() which causes error
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	cfg.AuthKey = "user"
	// parsing flags causes flag redefined errors
	//cfg.ParseFlags()
	suite.ctx, suite.cancel = context.WithCancel(context.Background())
	suite.wg = &sync.WaitGroup{}
	suite.wg.Add(1)
	suite.storage, _ = infile.InitStorage(suite.ctx, suite.wg, cfg)
	suite.shortenerService, _ = shortener.InitShortener(suite.storage)
	suite.urlHandler, _ = InitURLHandler(suite.shortenerService, cfg)
	suite.secretaryService = secretary.NewSecretaryService(cfg)
	suite.cookieHandler, _ = middleware.NewCookieHandler(suite.secretaryService, cfg)
	suite.router = chi.NewRouter()
	suite.ts = httptest.NewServer(suite.router)
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

func (suite *HandlersTestSuite) TestHandleGetStats() {
	userID := suite.secretaryService.Encode(uuid.New().String())
	_, _ = suite.shortenerService.Encode(suite.ctx, "https://www.yandex.ru", userID)
	_, _ = suite.shortenerService.Encode(suite.ctx, "https://www.yandex.com", userID)
	_, _ = suite.shortenerService.Encode(suite.ctx, "https://www.yandex.kz", userID)
	suite.router.Get("/api/internal/stats", suite.urlHandler.HandleGetStats())

	// set tests' parameters
	type want struct {
		code  int
		value modeldto.ResponseStats
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "Sample query",
			want: want{
				code: 200,
				value: modeldto.ResponseStats{
					URLs:  3,
					Users: 1,
				},
			},
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			client := resty.New()

			res, err := client.R().Get(suite.ts.URL + "/api/internal/stats")
			if err != nil {
				t.Fatalf(err.Error())
			}
			assert.Equal(t, tt.want.code, res.StatusCode())
			rb := res.Body()
			var expectedResponse modeldto.ResponseStats
			_ = json.Unmarshal(rb, &expectedResponse)
			assert.Equal(t, tt.want.value, expectedResponse)
		})
	}
	defer suite.ts.Close()
	suite.cancel()
	suite.wg.Wait()
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
		{
			name: "Absent GET query",
			sURL: "some_absent_sURL",
			want: want{
				code: 400,
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
				t.Log(err)
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

// Benchmarks

func BenchmarkInitURLHandler(b *testing.B) {
	cfg := config.NewDefaultConfiguration()
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	svc, _ := shortener.InitShortener(strg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = InitURLHandler(svc, cfg)
	}
}

func BenchmarkURLHandler_HandleGetURL(b *testing.B) {
	cfg := config.NewDefaultConfiguration()
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	svc, _ := shortener.InitShortener(strg)
	urlHandler, _ := InitURLHandler(svc, cfg)
	secretaryService := secretary.NewSecretaryService(cfg)
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
	cfg := config.NewDefaultConfiguration()
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	svc, _ := shortener.InitShortener(strg)
	urlHandler, _ := InitURLHandler(svc, cfg)
	secretaryService := secretary.NewSecretaryService(cfg)
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg)
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
	cfg := config.NewDefaultConfiguration()
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	svc, _ := shortener.InitShortener(strg)
	urlHandler, _ := InitURLHandler(svc, cfg)
	secretaryService := secretary.NewSecretaryService(cfg)
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg)
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
	cfg := config.NewDefaultConfiguration()
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	svc, _ := shortener.InitShortener(strg)
	urlHandler, _ := InitURLHandler(svc, cfg)
	secretaryService := secretary.NewSecretaryService(cfg)
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg)
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
	cfg := config.NewDefaultConfiguration()
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	svc, _ := shortener.InitShortener(strg)
	urlHandler, _ := InitURLHandler(svc, cfg)
	secretaryService := secretary.NewSecretaryService(cfg)
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg)
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
	cfg := config.NewDefaultConfiguration()
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	svc, _ := shortener.InitShortener(strg)
	urlHandler, _ := InitURLHandler(svc, cfg)
	secretaryService := secretary.NewSecretaryService(cfg)
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg)
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
	cfg := config.NewDefaultConfiguration()
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	svc, _ := shortener.InitShortener(strg)
	urlHandler, _ := InitURLHandler(svc, cfg)
	secretaryService := secretary.NewSecretaryService(cfg)
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg)
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	router.Use(cookieHandler.CookieHandle)
	router.Delete("/api/user/urls", urlHandler.HandleDeleteURLBatch())
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
			_, _ = client.R().SetBody(payload).Delete(ts.URL + "/api/user/urls")
		}
	})
	cancel()
	wg.Wait()
}

// Examples

func ExampleInitURLHandler() {
	// Parse environment
	cfg := config.NewDefaultConfiguration()
	// Parse CLI-defined flags and arguments in a MWE, not in tests
	//cfg.ParseFlags()
	// Set parameters explicitly for error-prone example running
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	cfg.AuthKey = "user"
	// Add context and wait group for storage operation control
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	// Initialize storage
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	// Initialize shortener service
	svc, _ := shortener.InitShortener(strg)
	// Initialize URL handler
	urlHandler, _ := InitURLHandler(svc, cfg)
	// Initialize router
	router := chi.NewRouter()
	// Set any available endpoint handler to any custom endpoint
	router.Get("/{urlID}", urlHandler.HandleGetURL())
	// Cancel context and wait for safe storage closure
	cancel()
	wg.Wait()
}

func ExampleURLHandler_HandleGetURL() {
	// Parse environment
	cfg := config.NewDefaultConfiguration()
	// Parse CLI-defined flags and arguments in a MWE, not in tests
	//cfg.ParseFlags()
	// Set parameters explicitly for error-prone example running
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	cfg.AuthKey = "user"
	// Add context and wait group for storage operation control
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	// Initialize storage
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	// Initialize shortener service
	svc, _ := shortener.InitShortener(strg)
	// Initialize URL handler
	urlHandler, _ := InitURLHandler(svc, cfg)
	// Initialize router
	router := chi.NewRouter()
	// Initialize secretary service
	secretaryService := secretary.NewSecretaryService(cfg)
	//// Initialize cookie handler
	//cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg)
	// Initialize server
	ts := httptest.NewServer(router)
	defer ts.Close()
	// Set a handler to an endpoint
	router.Get("/{urlID}", urlHandler.HandleGetURL())
	// Prepare test data
	userID := secretaryService.Encode(uuid.New().String())
	sURL, _ := svc.Encode(ctx, "https://www.example-url-1.com", userID)
	// Create a new client
	client := resty.New()
	client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}))
	// Execute a query
	res, _ := client.R().SetPathParams(map[string]string{"urlID": sURL}).Get(ts.URL + "/{urlID}")
	fmt.Println(res.StatusCode())
	cancel()
	wg.Wait()

	// Output:
	// 307
}

func ExampleURLHandler_HandlePostURL() {
	// Parse environment
	cfg := config.NewDefaultConfiguration()
	// Parse CLI-defined flags and arguments in a MWE, not in tests
	//cfg.ParseFlags()
	// Set parameters explicitly for error-prone example running
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	cfg.AuthKey = "user"
	// Add context and wait group for storage operation control
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	// Initialize storage
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	// Initialize shortener service
	svc, _ := shortener.InitShortener(strg)
	// Initialize URL handler
	urlHandler, _ := InitURLHandler(svc, cfg)
	// Initialize router
	router := chi.NewRouter()
	// Initialize secretary service
	secretaryService := secretary.NewSecretaryService(cfg)
	// Initialize cookie handler
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg)
	// Initialize server
	ts := httptest.NewServer(router)
	defer ts.Close()
	// Add authorization middleware via cookies
	router.Use(cookieHandler.CookieHandle)
	// Set a handler to an endpoint
	router.Post("/", urlHandler.HandlePostURL())
	// Create a query payload
	payload := strings.NewReader("https://www.sexample-url-2.com")
	// Create a new client
	client := resty.New()
	// Execute a query
	res, _ := client.R().SetBody(payload).Post(ts.URL)
	fmt.Println(res.StatusCode())
	cancel()
	wg.Wait()

	// Output:
	// 201
}

func ExampleURLHandler_JSONHandlePostURL() {
	// Parse environment
	cfg := config.NewDefaultConfiguration()
	// Parse CLI-defined flags and arguments in a MWE, not in tests
	//cfg.ParseFlags()
	// Set parameters explicitly for error-prone example running
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	cfg.AuthKey = "user"
	// Add context and wait group for storage operation control
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	// Initialize storage
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	// Initialize shortener service
	svc, _ := shortener.InitShortener(strg)
	// Initialize URL handler
	urlHandler, _ := InitURLHandler(svc, cfg)
	// Initialize router
	router := chi.NewRouter()
	// Initialize secretary service
	secretaryService := secretary.NewSecretaryService(cfg)
	// Initialize cookie handler
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg)
	// Initialize server
	ts := httptest.NewServer(router)
	defer ts.Close()
	// Add authorization middleware via cookies
	router.Use(cookieHandler.CookieHandle)
	// Set a handler to an endpoint
	router.Post("/api/shorten", urlHandler.JSONHandlePostURL())
	// Create a query payload
	URL := modeldto.RequestURL{
		URL: "https://www.example-url-3.com",
	}
	reqBody, _ := json.Marshal(URL)
	payload := strings.NewReader(string(reqBody))
	// Create a new client
	client := resty.New()
	// Execute a query
	res, _ := client.R().SetBody(payload).Post(ts.URL + "/api/shorten")
	fmt.Println(res.StatusCode())
	cancel()
	wg.Wait()

	// Output:
	// 201
}

func ExampleURLHandler_JSONHandlePostURLBatch() {
	// Parse environment
	cfg := config.NewDefaultConfiguration()
	// Parse CLI-defined flags and arguments in a MWE, not in tests
	//cfg.ParseFlags()
	// Set parameters explicitly for error-prone example running
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	cfg.AuthKey = "user"
	// Add context and wait group for storage operation control
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	// Initialize storage
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	// Initialize shortener service
	svc, _ := shortener.InitShortener(strg)
	// Initialize URL handler
	urlHandler, _ := InitURLHandler(svc, cfg)
	// Initialize router
	router := chi.NewRouter()
	// Initialize secretary service
	secretaryService := secretary.NewSecretaryService(cfg)
	// Initialize cookie handler
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg)
	// Initialize server
	ts := httptest.NewServer(router)
	defer ts.Close()
	// Add authorization middleware via cookies
	router.Use(cookieHandler.CookieHandle)
	// Set a handler to an endpoint
	router.Post("/api/shorten/batch", urlHandler.JSONHandlePostURLBatch())
	// Create a query payload
	batch := []modeldto.RequestBatchURL{
		{
			CorrelationID: "test1",
			URL:           "https://www.example-url-4.com",
		},
		{
			CorrelationID: "test2",
			URL:           "https://www.example-url-5.com",
		},
	}
	reqBody, _ := json.Marshal(batch)
	payload := strings.NewReader(string(reqBody))
	// Create a new client
	client := resty.New()
	// Execute a query
	res, _ := client.R().SetBody(payload).Post(ts.URL + "/api/shorten/batch")
	fmt.Println(res.StatusCode())
	cancel()
	wg.Wait()

	// Output:
	// 201
}

func ExampleURLHandler_HandleDeleteURLBatch() {
	// Parse environment
	cfg := config.NewDefaultConfiguration()
	// Parse CLI-defined flags and arguments in a MWE, not in tests
	//cfg.ParseFlags()
	// Set parameters explicitly for error-prone example running
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	cfg.AuthKey = "user"
	// Add context and wait group for storage operation control
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	// Initialize storage
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	// Initialize shortener service
	svc, _ := shortener.InitShortener(strg)
	// Initialize URL handler
	urlHandler, _ := InitURLHandler(svc, cfg)
	// Initialize router
	router := chi.NewRouter()
	// Initialize secretary service
	secretaryService := secretary.NewSecretaryService(cfg)
	// Initialize cookie handler
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg)
	// Initialize server
	ts := httptest.NewServer(router)
	defer ts.Close()
	// Add authorization middleware via cookies
	router.Use(cookieHandler.CookieHandle)
	// Set a handler to an endpoint
	router.Delete("/api/user/urls", urlHandler.HandleDeleteURLBatch())
	// Create a query payload
	batch := []string{"235n3g563jh5v3", "234g2342h5423"}
	reqBody, _ := json.Marshal(batch)
	payload := strings.NewReader(string(reqBody))
	// Create a new client
	client := resty.New()
	// Execute a query
	res, _ := client.R().SetBody(payload).Delete(ts.URL + "/api/user/urls")
	fmt.Println(res.StatusCode())
	cancel()
	wg.Wait()

	// Output:
	// 202
}

func ExampleURLHandler_HandleGetURLsByUserID() {
	// Parse environment
	cfg := config.NewDefaultConfiguration()
	// Parse CLI-defined flags and arguments in a MWE, not in tests
	//cfg.ParseFlags()
	// Set parameters explicitly for error-prone example running
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	cfg.AuthKey = "user"
	// Add context and wait group for storage operation control
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	// Initialize storage
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	// Initialize shortener service
	svc, _ := shortener.InitShortener(strg)
	// Initialize URL handler
	urlHandler, _ := InitURLHandler(svc, cfg)
	// Initialize router
	router := chi.NewRouter()
	// Initialize secretary service
	secretaryService := secretary.NewSecretaryService(cfg)
	// Initialize cookie handler
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg)
	// Initialize server
	ts := httptest.NewServer(router)
	defer ts.Close()
	// Add authorization middleware via cookies
	router.Use(cookieHandler.CookieHandle)
	// Set a handler to an endpoint
	router.Get("/api/user/urls", urlHandler.HandleGetURLsByUserID())
	// prepare test data
	userIDFull := secretaryService.Encode(uuid.New().String())
	_, _ = svc.Encode(ctx, "https://www.example-url-6.com", userIDFull)
	// Create a new client
	client := resty.New()
	client.SetCookie(&http.Cookie{
		Name:  "user",
		Value: userIDFull,
		Path:  "/",
	})
	// Execute a query
	res, _ := client.R().Get(ts.URL + "/api/user/urls")
	fmt.Println(res.StatusCode())
	cancel()
	wg.Wait()

	// Output:
	// 200
}

func ExampleURLHandler_HandlePingDB() {
	// Parse environment
	cfg := config.NewDefaultConfiguration()
	// Parse CLI-defined flags and arguments in a MWE, not in tests
	//cfg.ParseFlags()
	// Set parameters explicitly for error-prone example running
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	cfg.AuthKey = "user"
	// Add context and wait group for storage operation control
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	// Initialize storage
	strg, _ := infile.InitStorage(ctx, wg, cfg)
	// Initialize shortener service
	svc, _ := shortener.InitShortener(strg)
	// Initialize URL handler
	urlHandler, _ := InitURLHandler(svc, cfg)
	// Initialize router
	router := chi.NewRouter()
	// Initialize secretary service
	secretaryService := secretary.NewSecretaryService(cfg)
	// Initialize cookie handler
	cookieHandler, _ := middleware.NewCookieHandler(secretaryService, cfg)
	// Initialize server
	ts := httptest.NewServer(router)
	defer ts.Close()
	// Add authorization middleware via cookies
	router.Use(cookieHandler.CookieHandle)
	// Set a handler to an endpoint
	router.Get("/ping", urlHandler.HandlePingDB())
	// Create a new client
	client := resty.New()
	res, _ := client.R().Get(ts.URL + "/ping")
	fmt.Println(res.StatusCode())
	cancel()
	wg.Wait()

	// Output:
	// 200
}
