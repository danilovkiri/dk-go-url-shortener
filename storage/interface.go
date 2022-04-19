// Package storage provides interfaces for types to be in compliance with.
package storage

import "context"

// URLSetter defines a set of methods for types implementing URLSetter.
type URLSetter interface {
	Dump(ctx context.Context, URL string, sURL string) error
}

// URLGetter defines a set of methods for types implementing URLGetter.
type URLGetter interface {
	Retrieve(ctx context.Context, sURL string) (URL string, err error)
}

// URLStorage defines a set of embedded interfaces for types implementing URLStorage.
type URLStorage interface {
	URLSetter
	URLGetter
}
