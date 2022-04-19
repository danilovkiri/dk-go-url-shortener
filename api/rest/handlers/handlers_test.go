package handlers

import (
	"context"
	shortenerService "github.com/danilovkiri/dk_go_url_shortener/service/shortener"
	"github.com/danilovkiri/dk_go_url_shortener/service/shortener/v1"
	"github.com/danilovkiri/dk_go_url_shortener/storage"
	"github.com/danilovkiri/dk_go_url_shortener/storage/inmemory"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type HandlersTestSuite struct {
	suite.Suite
	storage          storage.URLStorage
	shortenerService shortenerService.Processor
	urlHandler       *URLHandler
}

func (suite *HandlersTestSuite) SetupTest() {
	suite.storage = inmemory.InitStorage()
	suite.shortenerService, _ = shortener.InitShortener(suite.storage)
	suite.urlHandler, _ = InitURLHandler(suite.shortenerService)
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

func (suite *HandlersTestSuite) TestHandleGetURL() {
	ctx := context.Background()
	sURL, _ := suite.shortenerService.Encode(ctx, "https://yandex.ru")
	r := chi.NewRouter()
	r.Get("/{urlID}", suite.urlHandler.HandleGetURL())
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

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(r)
			defer ts.Close()
			req, err := http.NewRequest(http.MethodGet, ts.URL+"/"+tt.sURL, nil)
			if err != nil {
				t.Fatalf("Could not create GET request")
			}
			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			res, err := client.Do(req)
			if err != nil {
				t.Fatalf(err.Error())
			} else {
				if res.StatusCode != tt.want.code {
					t.Fatalf("Expected status code %d, got %d", tt.want.code, res.StatusCode)
				}
			}

		})
	}
}

func (suite *HandlersTestSuite) TestHandlePostURL() {
	r := chi.NewRouter()
	r.Post("/", suite.urlHandler.HandlePostURL())
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
			URL:  "https://www.yandex.ru",
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

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(r)
			defer ts.Close()
			payload := strings.NewReader(tt.URL)
			req, err := http.NewRequest(http.MethodPost, ts.URL+"/", payload)
			if err != nil {
				t.Fatalf("Could not create POST request")
			}
			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				t.Errorf(err.Error())
			}
			assert.Equal(t, tt.want.code, res.StatusCode)

		})
	}
}
