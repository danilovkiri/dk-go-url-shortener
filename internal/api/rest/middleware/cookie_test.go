package middleware

import (
	"errors"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/mocks"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCookieHandleAbsentCookie(t *testing.T) {
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	cfg, _ := config.NewSecretConfig()
	cfg.UserKey = "jds__63h3_7ds"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mocks.NewMockSecretary(ctrl)
	cookieHandler, _ := NewCookieHandler(s, cfg)
	router.Use(cookieHandler.CookieHandle)
	router.Get("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("authorized"))
	})
	requestCookie := &http.Cookie{
		Name:  "some-other-key",
		Value: "some-token",
		Raw:   "user=some-token; Path=/",
		Path:  "/",
	}
	responseCookie := &http.Cookie{
		Name:  UserCookieKey,
		Value: "some-expected-token",
		Raw:   "user=some-expected-token; Path=/",
		Path:  "/",
	}
	s.EXPECT().Encode(gomock.Any()).Return(responseCookie.Value)
	client := resty.New()
	res, err := client.R().SetCookie(requestCookie).Get(ts.URL + "/get")
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, 200, res.StatusCode())
	assert.Equal(t, responseCookie, res.Cookies()[0])
}

func TestCookieHandleGoodCookie(t *testing.T) {
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	cfg, _ := config.NewSecretConfig()
	cfg.UserKey = "jds__63h3_7ds"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mocks.NewMockSecretary(ctrl)
	cookieHandler, _ := NewCookieHandler(s, cfg)
	router.Use(cookieHandler.CookieHandle)
	router.Get("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("authorized"))
	})
	requestCookie := &http.Cookie{
		Name:  UserCookieKey,
		Value: "some-expected-token",
		Raw:   "user=some-expected-token; Path=/",
		Path:  "/",
	}
	s.EXPECT().Decode(gomock.Any()).Return("some-expected-token-deciphered", nil)
	client := resty.New()
	res, err := client.R().SetCookie(requestCookie).Get(ts.URL + "/get")
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, 200, res.StatusCode())
}

func TestCookieHandleBadCookie(t *testing.T) {
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	cfg, _ := config.NewSecretConfig()
	cfg.UserKey = "jds__63h3_7ds"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mocks.NewMockSecretary(ctrl)
	cookieHandler, _ := NewCookieHandler(s, cfg)
	router.Use(cookieHandler.CookieHandle)
	router.Get("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("authorized"))
	})
	requestCookie := &http.Cookie{
		Name:  UserCookieKey,
		Value: "some-erroneous-token",
		Raw:   "user=some-erroneous-token; Path=/",
		Path:  "/",
	}
	s.EXPECT().Decode(gomock.Any()).Return("", errors.New("some-generic-error"))
	client := resty.New()
	res, err := client.R().SetCookie(requestCookie).Get(ts.URL + "/get")
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, 401, res.StatusCode())
}
