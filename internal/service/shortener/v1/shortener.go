// Package shortener provides functionality for creating a short unique identifier for a string.
package shortener

import (
	"context"
	"net/url"
	"time"

	"github.com/speps/go-hashids/v2"

	serviceErrors "github.com/danilovkiri/dk_go_url_shortener/internal/service/errors"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/modelurl"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/shortener"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/modelstorage"
)

const SaltKey = "Some Hashing Key"
const MinLength = 5

// Check interface implementation explicitly
var (
	_ shortener.Processor = (*Shortener)(nil)
)

// Shortener struct defines data structure handling and provides support for adding new implementations.
type Shortener struct {
	SaltKey    string
	MinLength  int
	hashID     *hashids.HashID
	URLStorage storage.URLStorage
}

// InitShortener initializes a Shortener object and sets its attributes.
func InitShortener(s storage.URLStorage) (*Shortener, error) {
	if s == nil {
		return nil, &serviceErrors.ServiceFoundNilStorage{Msg: "nil storage was passed to service initializer"}
	}
	hd := hashids.NewData()
	hd.Salt = SaltKey
	hd.MinLength = MinLength
	hashID, err := hashids.NewWithData(hd)
	if err != nil {
		return nil, &serviceErrors.ServiceInitHashError{Msg: err.Error()}
	}
	shortener := &Shortener{
		SaltKey:    SaltKey,
		MinLength:  MinLength,
		hashID:     hashID,
		URLStorage: s,
	}
	return shortener, nil
}

// Encode generates a sURL, stores URL and sURL in a storage, and returns sURL.
func (short *Shortener) Encode(ctx context.Context, URL string, userID string) (sURL string, err error) {
	_, err = url.ParseRequestURI(URL)
	if err != nil {
		return "", &serviceErrors.ServiceIncorrectInputURL{Msg: err.Error()}
	}
	sURL, err = short.generateSlug()
	if err != nil {
		return "", &serviceErrors.ServiceEncodingHashError{Msg: err.Error()}
	}
	err = short.URLStorage.Dump(ctx, URL, sURL, userID)
	if err != nil {
		return "", err
	}
	return sURL, nil
}

// Decode retrieves and returns URL based on the given sURL as a key.
func (short *Shortener) Decode(ctx context.Context, sURL string) (URL string, err error) {
	URL, err = short.URLStorage.Retrieve(ctx, sURL)
	if err != nil {
		return "", err
	}
	return URL, nil
}

// Delete performs soft removal of URL-sURL entries with task management and resource allocation.
func (short *Shortener) Delete(ctx context.Context, sURLs []string, userID string) {
	for i := 0; i < len(sURLs); i++ {
		item := modelstorage.URLChannelEntry{UserID: userID, SURL: sURLs[i]}
		short.URLStorage.SendToQueue(item)
	}
}

// DecodeByUserID retrieves and returns all pairs of sURL:URL for a given user ID.
func (short *Shortener) DecodeByUserID(ctx context.Context, userID string) (URLs []modelurl.FullURL, err error) {
	URLs, err = short.URLStorage.RetrieveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return URLs, nil
}

func (short *Shortener) PingDB() error {
	err := short.URLStorage.PingDB()
	return err
}

// generateSlug generates and returns a short unique identifier for a string.
func (short *Shortener) generateSlug() (slug string, err error) {
	now := time.Now().UnixNano()
	slug, err = short.hashID.Encode([]int{int(now)})
	return slug, err
}
