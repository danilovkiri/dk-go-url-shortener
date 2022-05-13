package infile

import (
	"context"
	"encoding/json"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/errors"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/modelstorage"
	"log"
	"os"
	"sync"
)

// Storage struct defines data structure handling and provides support for adding new implementations.
type Storage struct {
	mu  sync.Mutex
	Cfg *config.StorageConfig
	DB  map[string]string
}

// InitStorage initializes a Storage object, sets its attributes and starts a listener for persistStorage.
func InitStorage(ctx context.Context, wg *sync.WaitGroup, cfg *config.StorageConfig) (*Storage, error) {
	db := make(map[string]string)
	st := Storage{
		DB:  db,
		Cfg: cfg,
	}
	err := st.restore()
	if err != nil {
		return nil, err
	}
	// start a goroutine to listen for ctx cancellation followed by persistStorage call
	// use sync.WaitGroup to prevent goroutine premature termination when main exits
	go func() {
		defer wg.Done()
		<-ctx.Done()
		err := st.persistStorage()
		if err != nil {
			log.Fatal(err)
		}
	}()
	return &st, nil
}

// Retrieve returns a URL as a value of a map based on the given sURL as a key of a map.
func (s *Storage) Retrieve(ctx context.Context, sURL string) (URL string, err error) {
	// create channels for listening to the go routine result
	retrieveDone := make(chan string)
	retrieveError := make(chan string)
	go func() {
		s.mu.Lock()
		URL, ok := s.DB[sURL]
		s.mu.Unlock()
		if !ok {
			retrieveError <- "not found in DB"
			return
		}
		retrieveDone <- URL
	}()

	// wait for the first channel to retrieve a value
	select {
	case <-ctx.Done():
		log.Println("Retrieving URL:", ctx.Err())
		return "", errors.ContextTimeoutExceededError{}
	case errString := <-retrieveError:
		log.Println("Retrieving URL:", errString)
		return "", errors.StorageNotFoundError{ID: sURL}
	case URL := <-retrieveDone:
		log.Println("Retrieving URL:", sURL, "as", URL)
		return URL, nil
	}
}

// Dump stores a pair of sURL and URL as a key-value pair in a map.
func (s *Storage) Dump(ctx context.Context, URL string, sURL string) error {
	// create channels for listening to the go routine result
	dumpDone := make(chan bool)
	dumpError := make(chan string)
	go func() {
		_, ok := s.DB[sURL]
		if ok {
			dumpError <- "already exists in DB"
			return
		}
		s.mu.Lock()
		s.DB[sURL] = URL
		s.mu.Unlock()
		dumpDone <- true
	}()

	// wait for the first channel to retrieve a value
	select {
	case <-ctx.Done():
		log.Println("Dumping URL:", ctx.Err())
		return errors.ContextTimeoutExceededError{}
	case errString := <-dumpError:
		log.Println("Dumping URL:", errString)
		return errors.StorageAlreadyExistsError{ID: sURL}
	case <-dumpDone:
		log.Println("Dumping URL:", sURL, "as", URL)
		return nil
	}
}

// restore fills the tmpfs DB with URL-sURL entries from file storage.
func (s *Storage) restore() error {
	var storageEntries []modelstorage.URLStorageEntry
	file, err := os.OpenFile(s.Cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&storageEntries)
	if err != nil {
		// decoding an empty file results in EOF error
		log.Println("Attempted to restore DB from empty file")
	}
	log.Print("DB was restored")
	for _, entry := range storageEntries {
		s.DB[entry.SURL] = entry.URL
	}
	return nil
}

// persistStorage appends the tmpfs DB contents to a file storage.
func (s *Storage) persistStorage() error {
	if len(s.DB) == 0 {
		log.Print("Empty DB to be saved")
		return nil
	}
	file, err := os.OpenFile(s.Cfg.FileStoragePath, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	var rows []modelstorage.URLStorageEntry
	for sURL, URL := range s.DB {
		rowToEncode := modelstorage.URLStorageEntry{
			SURL: sURL,
			URL:  URL,
		}
		rows = append(rows, rowToEncode)
	}
	err = encoder.Encode(rows)
	if err != nil {
		return err
	}
	log.Print("DB was saved")
	return nil
}
