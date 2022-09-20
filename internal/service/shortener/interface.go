// Package shortener provides interfaces for types to be in compliance with.
package shortener

import (
	"context"

	"github.com/danilovkiri/dk_go_url_shortener/internal/service/modelurl"
)

// Processor defines a set of methods for types implementing Processor.
type Processor interface {
	GetStats(ctx context.Context) (nURLs, nUsers int64, err error)
	Encode(ctx context.Context, URL, userID string) (sURL string, err error)
	Decode(ctx context.Context, sURL string) (URL string, err error)
	Delete(ctx context.Context, sURLs []string, userID string)
	DecodeByUserID(ctx context.Context, userID string) (URLs []modelurl.FullURL, err error)
	PingDB() error
}
