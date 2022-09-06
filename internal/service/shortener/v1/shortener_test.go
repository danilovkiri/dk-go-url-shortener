package shortener

import (
	"context"
	"errors"
	"testing"

	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/mocks"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/modelurl"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/modelstorage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// Tests

func TestInitShortener(t *testing.T) {
	cfg := config.NewDefaultConfiguration()
	// necessary to set default parameters here since they are set in cfg.ParseFlags() which causes error
	cfg.ServerAddress = ":8080"
	cfg.BaseURL = "http://localhost:8080"
	cfg.FileStoragePath = "url_storage.json"
	_, err := InitShortener(nil)
	assert.Equal(t, "nil storage was passed to service initializer", err.Error())
}

func TestPingDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mocks.NewMockURLStorage(ctrl)
	s.EXPECT().PingDB().Return(nil)
	processor, _ := InitShortener(s)
	processor.PingDB()
}

func TestDecodeByUserID_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mocks.NewMockURLStorage(ctrl)
	userID := "someUserID"
	s.EXPECT().RetrieveByUserID(context.Background(), userID).Return(nil, errors.New("generic error"))
	processor, _ := InitShortener(s)
	_, err := processor.DecodeByUserID(context.Background(), userID)
	assert.Equal(t, errors.New("generic error"), err)
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

func TestDecode_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mocks.NewMockURLStorage(ctrl)
	sURL := "someShortURL"
	s.EXPECT().Retrieve(context.Background(), sURL).Return("", errors.New("generic error"))
	processor, _ := InitShortener(s)
	_, err := processor.Decode(context.Background(), sURL)
	assert.Equal(t, errors.New("generic error"), err)
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
func TestEncode_Fail1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mocks.NewMockURLStorage(ctrl)
	URL := "some_invalid_URL"
	userID := "someUserID"
	processor, _ := InitShortener(s)
	_, err := processor.Encode(context.Background(), URL, userID)
	assert.Equal(t, "parse \"some_invalid_URL\": invalid URI for request", err.Error())
}

func TestEncode_Fail2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	s := mocks.NewMockURLStorage(ctrl)
	URL := "https://www.some-url.com"
	userID := "someUserID"
	s.EXPECT().Dump(context.Background(), URL, gomock.Any(), userID).Return(errors.New("generic error"))
	processor, _ := InitShortener(s)
	_, err := processor.Encode(context.Background(), URL, userID)
	assert.Equal(t, errors.New("generic error"), err)
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

// Benchmarks

func BenchmarkInitShortener(b *testing.B) {
	ctrl := gomock.NewController(b)
	s := mocks.NewMockURLStorage(ctrl)
	defer ctrl.Finish()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = InitShortener(s)
	}
}

func BenchmarkShortener_PingDB(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	s := mocks.NewMockURLStorage(ctrl)
	s.EXPECT().PingDB().Return(nil).AnyTimes()
	processor, _ := InitShortener(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.PingDB()
	}
}

func BenchmarkShortener_Encode(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	s := mocks.NewMockURLStorage(ctrl)
	URL := "https://www.some-url.com"
	userID := "someUserID"
	s.EXPECT().Dump(context.Background(), URL, gomock.Any(), userID).Return(nil).AnyTimes()
	processor, _ := InitShortener(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = processor.Encode(context.Background(), URL, userID)
	}
}

func BenchmarkShortener_Decode(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	s := mocks.NewMockURLStorage(ctrl)
	sURL := "someShortURL"
	URL := "someURL"
	s.EXPECT().Retrieve(context.Background(), sURL).Return(URL, nil).AnyTimes()
	processor, _ := InitShortener(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = processor.Decode(context.Background(), sURL)
	}
}

func BenchmarkShortener_DecodeByUserID(b *testing.B) {
	ctrl := gomock.NewController(b)
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
	s.EXPECT().RetrieveByUserID(context.Background(), userID).Return(URLs, nil).AnyTimes()
	processor, _ := InitShortener(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = processor.DecodeByUserID(context.Background(), userID)
	}
}

func BenchmarkShortener_Delete(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	s := mocks.NewMockURLStorage(ctrl)
	userID := "someUserID"
	sURL := "someShortURL"
	sURLs := []string{sURL}
	item := modelstorage.URLChannelEntry{UserID: userID, SURL: sURL}
	s.EXPECT().SendToQueue(item).Return().AnyTimes()
	processor, _ := InitShortener(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.Delete(context.Background(), sURLs, userID)
	}
}
