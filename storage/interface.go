package storage

import "context"

type URLSetter interface {
	Dump(ctx context.Context, URL string, sURL string) error
}

type URLGetter interface {
	Retrieve(ctx context.Context, sURL string) (URL string, err error)
}

type URLStorage interface {
	URLSetter
	URLGetter
}
