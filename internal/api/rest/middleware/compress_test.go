package middleware

import (
	"bytes"
	"compress/gzip"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type CompressTestSuite struct {
	suite.Suite
	router *chi.Mux
	ts     *httptest.Server
}

func (suite *CompressTestSuite) SetupTest() {
	suite.router = chi.NewRouter()
	suite.ts = httptest.NewServer(suite.router)
}

func TestCompressTestSuite(t *testing.T) {
	suite.Run(t, new(CompressTestSuite))
}

func (suite *CompressTestSuite) TestCompressHandle() {
	suite.router.Use(CompressHandle)
	suite.router.Get("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("textstring"))
	})

	tests := []struct {
		name              string
		expectedEncoding  string
		acceptedEncodings []string
	}{
		{
			name:              "no encoding",
			acceptedEncodings: nil,
			expectedEncoding:  "",
		},
		{
			name:              "gzip encoding",
			acceptedEncodings: []string{"gzip"},
			expectedEncoding:  "gzip",
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			client := resty.New()
			res, err := client.R().SetHeader("Accept-Encoding", strings.Join(tt.acceptedEncodings, ",")).Get(suite.ts.URL + "/get")
			if err != nil {
				t.Fatalf(err.Error())
			}
			assert.Equal(t, tt.expectedEncoding, res.Header().Get("Content-Encoding"))
		})
	}
	defer suite.ts.Close()
}

func (suite *CompressTestSuite) TestDecompressHandle() {
	suite.router.Use(DecompressHandle)
	suite.router.Post("/post", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		_, err = w.Write(b)
		if err != nil {
			log.Fatal(err)
		}
	})

	tests := []struct {
		name          string
		queryEncoding string
		payload       string
	}{
		{
			name:          "no encoding",
			queryEncoding: "",
			payload:       "some_data",
		},
		{
			name:          "gzip encoding",
			queryEncoding: "gzip",
			payload:       "some_other_data",
		},
	}

	// perform each test
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			client := resty.New()
			var payload string
			if tt.queryEncoding == "" {
				payload = tt.payload
			} else {
				var b bytes.Buffer
				gz := gzip.NewWriter(&b)
				if _, err := gz.Write([]byte(tt.payload)); err != nil {
					log.Fatal(err)
				}
				if err := gz.Close(); err != nil {
					log.Fatal(err)
				}
				payload = b.String()
			}
			res, err := client.R().SetHeader("Content-Encoding", tt.queryEncoding).SetBody(payload).Post(suite.ts.URL + "/post")
			if err != nil {
				t.Fatalf(err.Error())
			}
			resBody := string(res.Body())
			assert.Equal(t, tt.payload, resBody)
		})
	}
	defer suite.ts.Close()
}

func BenchmarkCompressHandle(b *testing.B) {
	router := chi.NewRouter()
	client := resty.New()
	ts := httptest.NewServer(router)
	defer ts.Close()
	router.Use(CompressHandle)
	router.Get("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("textstring"))
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.R().SetHeader("Accept-Encoding", "gzip").Get(ts.URL + "/get")
	}
}

func BenchmarkDecompressHandle(b *testing.B) {
	router := chi.NewRouter()
	client := resty.New()
	ts := httptest.NewServer(router)
	defer ts.Close()
	router.Use(DecompressHandle)
	router.Post("/post", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		_, err = w.Write(b)
		if err != nil {
			log.Fatal(err)
		}
	})
	var bts bytes.Buffer
	gz := gzip.NewWriter(&bts)
	if _, err := gz.Write([]byte("some_data")); err != nil {
		log.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		log.Fatal(err)
	}
	payload := bts.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.R().SetHeader("Content-Encoding", "gzip").SetBody(payload).Post(ts.URL + "/post")
	}
}
