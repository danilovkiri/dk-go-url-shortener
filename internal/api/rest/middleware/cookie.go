// Package middleware provides various middleware functionality.
package middleware

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	serviceErrors "github.com/danilovkiri/dk_go_url_shortener/internal/service/errors"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/secretary"
)

// CookieHandler sets object structure.
type CookieHandler struct {
	sec secretary.Secretary
	cfg *config.SecretConfig
}

// UserCookieKey sets a cookie key to be used in user identification.
const UserCookieKey = "user"

// NewCookieHandler initializes a new cookie handler.
func NewCookieHandler(sec secretary.Secretary, cfg *config.SecretConfig) (*CookieHandler, error) {
	if sec == nil {
		return nil, &serviceErrors.ServiceFoundNilStorage{Msg: "nil secretary was passed to service initializer"}
	}
	return &CookieHandler{
		sec: sec,
		cfg: cfg,
	}, nil
}

// CookieHandle provides cookie handling functionality.
func (c *CookieHandler) CookieHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(UserCookieKey)
		if errors.Is(err, http.ErrNoCookie) {
			userID := uuid.New().String()
			token := c.sec.Encode(userID)
			newCookie := &http.Cookie{
				Name:  UserCookieKey,
				Value: token,
				Path:  "/",
			}
			http.SetCookie(w, newCookie)
			r.AddCookie(newCookie)
		} else if err != nil {
			http.Error(w, "Cookie crumbled", http.StatusInternalServerError)
		} else {
			_, err := c.sec.Decode(cookie.Value)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			}
		}
		next.ServeHTTP(w, r)
	})
}
