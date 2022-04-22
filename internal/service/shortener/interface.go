// Package shortener provides interfaces for types to be in compliance with.
package shortener

import "context"

// Processor defines a set of methods for types implementing Processor.
type Processor interface {
	Encode(ctx context.Context, URL string) (sURL string, err error)
	Decode(ctx context.Context, sURL string) (URL string, err error)
}
