package shortener

import (
	"github.com/danilovkiri/dk_go_url_shortener/service/errors"
	"github.com/speps/go-hashids/v2"
)

const SaltKey = "Some Hashing Key"
const MinLength = 5

type Shortener struct {
	SaltKey   string
	MinLength int
	hashID    *hashids.HashID
}

func InitShortener() (*Shortener, error) {
	hd := hashids.NewData()
	hd.Salt = SaltKey
	hd.MinLength = MinLength
	hashID, err := hashids.NewWithData(hd)
	if err != nil {
		return nil, &errors.ServiceInitHashError{Msg: err.Error()}
	}
	short := &Shortener{
		SaltKey:   SaltKey,
		MinLength: MinLength,
		hashID:    hashID,
	}
	return short, nil
}

func (short *Shortener) Encode(index int) (string, error) {
	hash, err := short.hashID.Encode([]int{index})
	if err != nil {
		return "", &errors.ServiceEncodingHashError{Msg: err.Error()}
	}
	return hash, nil
}

func (short *Shortener) Decode(hash string) (int, error) {
	decoded, err := short.hashID.DecodeWithError(hash)
	if err != nil {
		return -1, &errors.ServiceDecodingHashError{Msg: err.Error()}
	}
	return decoded[0], nil
}
