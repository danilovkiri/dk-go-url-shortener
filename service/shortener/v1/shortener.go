package shortener

import (
	"context"
	"net/url"
	"time"

	"github.com/danilovkiri/dk_go_url_shortener/service/errors"
	"github.com/danilovkiri/dk_go_url_shortener/storage"
	"github.com/speps/go-hashids/v2"
)

const SaltKey = "Some Hashing Key"
const MinLength = 5

type Shortener struct {
	SaltKey    string
	MinLength  int
	hashID     *hashids.HashID
	URLStorage storage.URLStorage
}

func InitShortener(s storage.URLStorage) (*Shortener, error) {
	if s == nil {
		return nil, &errors.ServiceFoundNilStorage{Msg: "nil storage was passed to service initializer"}
	}
	hd := hashids.NewData()
	hd.Salt = SaltKey
	hd.MinLength = MinLength
	hashID, err := hashids.NewWithData(hd)
	if err != nil {
		return nil, &errors.ServiceInitHashError{Msg: err.Error()}
	}
	shortener := &Shortener{
		SaltKey:    SaltKey,
		MinLength:  MinLength,
		hashID:     hashID,
		URLStorage: s,
	}
	return shortener, nil
}

func (short *Shortener) Encode(ctx context.Context, URL string) (sURL string, err error) {
	_, err = url.ParseRequestURI(URL)
	if err != nil {
		return "", &errors.ServiceIncorrectInputURL{Msg: err.Error()}
	}
	sURL, err = short.generateSlug()
	if err != nil {
		return "", &errors.ServiceEncodingHashError{Msg: err.Error()}
	}
	err = short.URLStorage.Dump(ctx, URL, sURL)
	if err != nil {
		return "", err
	}
	return sURL, nil
}

func (short *Shortener) Decode(ctx context.Context, sURL string) (URL string, err error) {
	URL, err = short.URLStorage.Retrieve(ctx, sURL)
	if err != nil {
		return "", err
	}
	return URL, nil
}

func (short *Shortener) generateSlug() (slug string, err error) {
	now := time.Now().UnixNano()
	slug, err = short.hashID.Encode([]int{int(now)})
	return slug, err
}
