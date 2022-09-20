package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/go-chi/chi"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// Tests

func TestNewTrustedNetHandler_InvalidCIDR(t *testing.T) {
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	cfg := config.NewDefaultConfiguration()
	cfg.TrustedSubnet = ""
	trustedNetHandler := NewTrustedNetHandler(cfg)
	expectedNethandler := &TrustedNetHandler{
		Resolved: false,
		IP:       nil,
		IPNet:    nil,
	}
	assert.Equal(t, expectedNethandler, trustedNetHandler)
}

func TestNewTrustedNetHandler(t *testing.T) {
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	cfg := config.NewDefaultConfiguration()
	cfg.TrustedSubnet = "127.135.1.0/24"
	trustedNetHandler := NewTrustedNetHandler(cfg)
	mask := net.IPMask(net.ParseIP("255.255.255.0").To4())
	expectedNethandler := &TrustedNetHandler{
		Resolved: true,
		IP:       net.ParseIP("127.135.1.0").To16(),
		IPNet: &net.IPNet{
			IP:   net.ParseIP("127.135.1.0").To4(),
			Mask: mask,
		},
	}
	assert.Equal(t, expectedNethandler, trustedNetHandler)
}

func TestTrustedNetHandler_TrustedNetworkHandler1(t *testing.T) {
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	cfg := config.NewDefaultConfiguration()
	cfg.TrustedSubnet = "127.135.1.0/24"
	trustedNetHandler := NewTrustedNetHandler(cfg)
	router.Use(trustedNetHandler.TrustedNetworkHandler)
	router.Get("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("authorized"))
	})
	client := resty.New()
	res, err := client.R().SetHeader("X-Real-IP", "127.135.1.1").Get(ts.URL + "/get")
	if err != nil {
		t.Fatalf(err.Error())
	}
	assert.Equal(t, 200, res.StatusCode())
}

func TestTrustedNetHandler_TrustedNetworkHandler2(t *testing.T) {
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	cfg := config.NewDefaultConfiguration()
	cfg.TrustedSubnet = "127.135.1.0/24"
	trustedNetHandler := NewTrustedNetHandler(cfg)
	router.Use(trustedNetHandler.TrustedNetworkHandler)
	router.Get("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("authorized"))
	})
	client := resty.New()
	res, err := client.R().SetHeader("X-Forwarded-For", "127.135.1.1").Get(ts.URL + "/get")
	if err != nil {
		t.Fatalf(err.Error())
	}
	assert.Equal(t, 200, res.StatusCode())
}

func TestTrustedNetHandler_TrustedNetworkHandler3(t *testing.T) {
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	cfg := config.NewDefaultConfiguration()
	cfg.TrustedSubnet = "127.135.1.0/24"
	trustedNetHandler := NewTrustedNetHandler(cfg)
	router.Use(trustedNetHandler.TrustedNetworkHandler)
	router.Get("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("authorized"))
	})
	client := resty.New()
	res, err := client.R().Get(ts.URL + "/get")
	if err != nil {
		t.Fatalf(err.Error())
	}
	assert.Equal(t, 403, res.StatusCode())
}

func TestTrustedNetHandler_TrustedNetworkHandler4(t *testing.T) {
	router := chi.NewRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	cfg := config.NewDefaultConfiguration()
	cfg.TrustedSubnet = ""
	trustedNetHandler := NewTrustedNetHandler(cfg)
	router.Use(trustedNetHandler.TrustedNetworkHandler)
	router.Get("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("authorized"))
	})
	client := resty.New()
	res, err := client.R().SetHeader("X-Real-IP", "127.135.1.1").Get(ts.URL + "/get")
	if err != nil {
		t.Fatalf(err.Error())
	}
	assert.Equal(t, 403, res.StatusCode())
}
