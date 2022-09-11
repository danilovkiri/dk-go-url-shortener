// Package infile provides data types and methods for local file storage operations.
package infile

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/modelurl"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1"
	storageErrors "github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/errors"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/v1/modelstorage"
)

// Check interface implementation explicitly
var (
	_ storage.URLStorage = (*Storage)(nil)
)

// Storage struct defines data structure handling and provides support for adding new implementations.
type Storage struct {
	mu      sync.Mutex
	Cfg     *config.Config
	DB      map[string]modelstorage.URLMapEntry
	Encoder *json.Encoder
}

// InitStorage initializes a Storage object and sets its attributes.
func InitStorage(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config) (*Storage, error) {
	db := make(map[string]modelstorage.URLMapEntry)
	st := Storage{
		Cfg: cfg,
		DB:  db,
	}
	err := st.restore()
	if err != nil {
		log.Fatal(err)
	}
	// open file outside goroutine since this operation might not finish prior to encoding operations
	file, err := os.OpenFile(st.Cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Fatal(err)
	}
	// set an encoder
	st.Encoder = json.NewEncoder(file)
	// start a goroutine to listen for ctx cancellation followed by file storage closure,
	// use sync.WaitGroup to prevent goroutine premature termination when main exits
	go func() {
		defer wg.Done()
		<-ctx.Done()
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("File storage closed successfully")
	}()
	return &st, nil
}

func (s *Storage) GetStats(ctx context.Context) (nURLs, nUsers int, err error) {
	// create channels for listening to the go routine result
	retrieveDone := make(chan []int, 1)
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		countURLs := len(s.DB)
		uniqueUsers := map[string]bool{}
		for _, URL := range s.DB {
			uniqueUsers[URL.UserID] = true
		}
		countUsers := len(uniqueUsers)
		retrieveDone <- []int{countURLs, countUsers}
	}()

	// wait for the first channel to retrieve a value
	select {
	case <-ctx.Done():
		log.Println("Retrieving stats:", ctx.Err())
		return 0, 0, &storageErrors.ContextTimeoutExceededError{Err: ctx.Err()}
	case stats := <-retrieveDone:
		log.Println("Retrieving stats: done")
		return stats[0], stats[1], nil
	}
}

// Retrieve returns a URL corresponding to sURL.
func (s *Storage) Retrieve(ctx context.Context, sURL string) (URL string, err error) {
	// create channels for listening to the go routine result
	retrieveDone := make(chan string)
	retrieveError := make(chan error)
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		URLMapEntry, ok := s.DB[sURL]
		if !ok {
			retrieveError <- &storageErrors.NotFoundError{Err: nil, SURL: sURL}
			return
		}
		retrieveDone <- URLMapEntry.URL
	}()

	// wait for the first channel to retrieve a value
	select {
	case <-ctx.Done():
		log.Println("Retrieving URL:", ctx.Err())
		return "", &storageErrors.ContextTimeoutExceededError{Err: ctx.Err()}
	case rtrvError := <-retrieveError:
		log.Println("Retrieving URL:", rtrvError.Error())
		return "", rtrvError
	case URL := <-retrieveDone:
		log.Println("Retrieving URL:", sURL, "as", URL)
		return URL, nil
	}
}

// RetrieveByUserID returns a slice of URL:sURL pairs defined as modelurl.FullURL for one particular user ID.
func (s *Storage) RetrieveByUserID(ctx context.Context, userID string) (URLs []modelurl.FullURL, err error) {
	// create channels for listening to the go routine result
	retrieveDone := make(chan []modelurl.FullURL)
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		var URLs []modelurl.FullURL
		for sURL, URL := range s.DB {
			if URL.UserID == userID {
				fullURL := modelurl.FullURL{
					URL:  URL.URL,
					SURL: sURL,
				}
				URLs = append(URLs, fullURL)
			}
		}
		retrieveDone <- URLs
	}()

	// wait for the first channel to retrieve a value
	select {
	case <-ctx.Done():
		log.Println("Retrieving URLs by UserID:", ctx.Err())
		return nil, &storageErrors.ContextTimeoutExceededError{Err: ctx.Err()}
	case URLs := <-retrieveDone:
		log.Println("Retrieving URL by UserID:", URLs)
		return URLs, nil
	}
}

// Dump stores a pair of sURL and URL as a key-value pair.
func (s *Storage) Dump(ctx context.Context, URL string, sURL string, userID string) error {
	// create channels for listening to the go routine result
	dumpDone := make(chan bool)
	dumpError := make(chan error)
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		_, ok := s.DB[sURL]
		if ok {
			dumpError <- &storageErrors.AlreadyExistsError{Err: nil, URL: sURL, ValidSURL: ""}
			return
		}
		s.DB[sURL] = modelstorage.URLMapEntry{URL: URL, UserID: userID}
		err := s.addToFileDB(sURL, URL, userID)
		if err != nil {
			dumpError <- &storageErrors.FileWriteError{Err: err}
			return
		}
		dumpDone <- true
	}()

	// wait for the first channel to retrieve a value
	select {
	case <-ctx.Done():
		log.Println("Dumping URL:", ctx.Err())
		return &storageErrors.ContextTimeoutExceededError{Err: ctx.Err()}
	case dmpError := <-dumpError:
		log.Println("Dumping URL:", dmpError.Error())
		return dmpError
	case <-dumpDone:
		log.Println("Dumping URL:", sURL, "as", URL)
		return nil
	}
}

// DeleteBatch is a mock for PSQL DB batch deleter for infile DB handling.
func (s *Storage) DeleteBatch(ctx context.Context, sURLs []string, userID string) error {
	return nil
}

// SendToQueue is a mock for PSQL DB batch concurrent deleter for infile DB handling.
func (s *Storage) SendToQueue(item modelstorage.URLChannelEntry) {
}

// restore fills the tmpfs DB with URL-sURL entries from file storage.
func (s *Storage) restore() error {
	var storageEntries []modelstorage.URLStorageEntry
	file, err := os.OpenFile(s.Cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bufio.NewScanner(file)
	for reader.Scan() {
		var storageEntry modelstorage.URLStorageEntry
		err := json.Unmarshal(reader.Bytes(), &storageEntry)
		if err != nil {
			return err
		}
		storageEntries = append(storageEntries, storageEntry)
	}
	log.Print("DB was restored")
	for _, entry := range storageEntries {
		s.DB[entry.SURL] = modelstorage.URLMapEntry{URL: entry.URL, UserID: entry.UserID}
	}
	return nil
}

// addToFileDB adds one sURL:URL key-value pair to a file DB.
func (s *Storage) addToFileDB(sURL, URL, userID string) error {
	rowToEncode := modelstorage.URLStorageEntry{
		SURL:   sURL,
		URL:    URL,
		UserID: userID,
	}
	err := s.Encoder.Encode(rowToEncode)
	if err != nil {
		return err
	}
	log.Print("POST query was saved to DB")
	return nil
}

// PingDB is a mock for PSQL DB pinger.
func (s *Storage) PingDB() error {
	return nil
}

// CloseDB is a mock for PSQL DB closer.
func (s *Storage) CloseDB() error {
	return nil
}
