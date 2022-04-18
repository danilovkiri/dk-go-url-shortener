package shortener

import "context"

type Processor interface {
	Encode(ctx context.Context, URL string) (sURL string, err error)
	Decode(ctx context.Context, sURL string) (URL string, err error)
}
