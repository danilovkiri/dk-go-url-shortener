package shortener

import (
	"context"
	"github.com/danilovkiri/dk_go_url_shortener/internal/mocks"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/modelurl"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/modelstorage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPingDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mocks.NewMockURLStorage(ctrl)
	s.EXPECT().PingDB().Return(nil)
	processor, _ := InitShortener(s)
	_ = processor.PingDB()
}

func TestDecodeByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mocks.NewMockURLStorage(ctrl)
	userID := "someUserID"
	URLs := []modelurl.FullURL{
		{
			URL:  "someURL1",
			SURL: "someShortURL1",
		},
		{
			URL:  "someURL2",
			SURL: "someShortURL2",
		},
	}
	s.EXPECT().RetrieveByUserID(context.Background(), userID).Return(URLs, nil)
	processor, _ := InitShortener(s)
	res, _ := processor.DecodeByUserID(context.Background(), userID)
	assert.Equal(t, URLs, res)
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mocks.NewMockURLStorage(ctrl)
	userID := "someUserID"
	sURL := "someShortURL"
	sURLs := []string{sURL}
	item := modelstorage.URLChannelEntry{UserID: userID, SURL: sURL}
	s.EXPECT().SendToQueue(item).Return()
	processor, _ := InitShortener(s)
	processor.Delete(context.Background(), sURLs, userID)
}

func TestDecode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mocks.NewMockURLStorage(ctrl)
	sURL := "someShortURL"
	URL := "someURL"
	s.EXPECT().Retrieve(context.Background(), sURL).Return(URL, nil)
	processor, _ := InitShortener(s)
	res, _ := processor.Decode(context.Background(), sURL)
	assert.Equal(t, URL, res)
}

func TestEncode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mocks.NewMockURLStorage(ctrl)
	URL := "https://www.some-url.com"
	userID := "someUserID"
	s.EXPECT().Dump(context.Background(), URL, gomock.Any(), userID).Return(nil)
	processor, _ := InitShortener(s)
	_, err := processor.Encode(context.Background(), URL, userID)
	assert.Equal(t, nil, err)
}
