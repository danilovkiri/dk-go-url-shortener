// Package middleware provides various middleware functionality.
package middleware

import (
	"errors"
	"net/http"

	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/secretary"
	"github.com/google/uuid"
)

// CookieHandler sets object structure.
type CookieHandler struct {
	sec secretary.Secretary
	cfg *config.Config
}

// NewCookieHandler initializes a new cookie handler.
func NewCookieHandler(sec secretary.Secretary, cfg *config.Config) (*CookieHandler, error) {
	return &CookieHandler{
		sec: sec,
		cfg: cfg,
	}, nil
}

// CookieHandle provides cookie handling functionality.
func (c *CookieHandler) CookieHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(c.cfg.AuthKey)
		if errors.Is(err, http.ErrNoCookie) {
			userID := uuid.New().String()
			token := c.sec.Encode(userID)
			newCookie := &http.Cookie{
				Name:  c.cfg.AuthKey,
				Value: token,
				Path:  "/",
			}
			http.SetCookie(w, newCookie)
			r.AddCookie(newCookie)
		} else {
			_, err := c.sec.Decode(cookie.Value)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			}
		}
		next.ServeHTTP(w, r)
	})
}
