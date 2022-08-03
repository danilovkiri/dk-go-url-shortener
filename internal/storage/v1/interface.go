// Package storage provides interfaces for types to be in compliance with.
package storage

import (
	"context"

	"github.com/danilovkiri/dk_go_url_shortener/internal/service/modelurl"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/modelstorage"
)

// URLSetter defines a set of methods for types implementing URLSetter.
type URLSetter interface {
	Dump(ctx context.Context, URL string, sURL string, userID string) error
}

// URLBatchDeleter defines a set of methods for types implementing URLBatchDeleter.
type URLBatchDeleter interface {
	DeleteBatch(ctx context.Context, sURLs []string, userID string) error
	SendToQueue(item modelstorage.URLChannelEntry)
}

// URLGetter defines a set of methods for types implementing URLGetter.
type URLGetter interface {
	Retrieve(ctx context.Context, sURL string) (URL string, err error)
}

// URLGetterByUserID defines a set of methods for types implementing URLGetterByUserID.
type URLGetterByUserID interface {
	RetrieveByUserID(ctx context.Context, userID string) (URLs []modelurl.FullURL, err error)
}

// Pinger defines a set of methods for types implementing Pinger.
type Pinger interface {
	PingDB() error
}

// Closer defines a set of methods for types implementing Closer.
type Closer interface {
	CloseDB() error
}

// URLStorage defines a set of embedded interfaces for types implementing URLStorage.
type URLStorage interface {
	URLSetter
	URLBatchDeleter
	URLGetter
	URLGetterByUserID
	Pinger
	Closer
}
