package inmemory

import (
	"github.com/danilovkiri/dk_go_url_shortener/storage/errors"
)

type (
	ShortUrl string
	Url      string
	Database struct {
		Storage map[ShortUrl]Url
	}
)

func InitStorage() *Database {
	db := &Database{Storage: make(map[ShortUrl]Url)}
	return db
}

func (db *Database) SaveShortUrl(sUrl ShortUrl, url Url) error {
	_, ok := db.Storage[sUrl]
	if !ok {
		db.Storage[sUrl] = url
		return nil
	}
	return &errors.StorageAlreadyExistsError{ShortURL: string(sUrl)}
}

func (db *Database) GetFullUrl(sUrl ShortUrl) (Url, error) {
	url, ok := db.Storage[sUrl]
	if !ok {
		return "", &errors.StorageNotFoundError{ShortURL: string(sUrl)}
	}
	return url, nil
}
