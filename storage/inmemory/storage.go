package inmemory

import (
	"github.com/danilovkiri/dk_go_url_shortener/storage/errors"
)

type (
	ShortUrl string
	Url      string
	LastUsed int
	Database struct {
		StorageF map[LastUsed]Url
		StorageR map[Url]LastUsed
		LastUsed
	}
)

func InitStorage() *Database {
	db := &Database{
		StorageF: make(map[LastUsed]Url),
		StorageR: make(map[Url]LastUsed),
		LastUsed: 0}
	return db
}

func (db *Database) Dump(url Url) (int, error) {
	_, ok := db.StorageR[url]
	if !ok {
		db.LastUsed++
		db.StorageF[db.LastUsed] = url
		db.StorageR[url] = db.LastUsed
		return int(db.LastUsed), nil
	}
	return int(db.LastUsed), &errors.StorageAlreadyExistsError{ID: string(url)}
}

func (db *Database) Retrieve(index int) (Url, error) {
	url, ok := db.StorageF[LastUsed(index)]
	if !ok {
		return "", &errors.StorageNotFoundError{ID: index}
	}
	return url, nil
}
