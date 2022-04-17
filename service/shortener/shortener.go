package shortener

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/danilovkiri/dk_go_url_shortener/service/errors"
)

func GenereteShortString(s string) (string, error) {
	h := md5.New()
	_, err := h.Write([]byte(s))
	if err != nil {
		return "", &errors.ServiceHashWriteError{Msg: err.Error()}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
