package infile

import (
	"bufio"
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
	mu      sync.Mutex
	Cfg     *config.StorageConfig
	DB      map[string]string
	Encoder *json.Encoder
}

// InitStorage initializes a Storage object, sets its attributes and starts a listener for persistStorage.
func InitStorage(ctx context.Context, wg *sync.WaitGroup, cfg *config.StorageConfig) (*Storage, error) {
	db := make(map[string]string)
	st := Storage{
		Cfg: cfg,
		DB:  db,
	}
	err := st.restore()
	if err != nil {
		return nil, err
	}
	// start a goroutine to open a file storage and set an Encoder object then
	// listen for ctx cancellation followed by file storage closure,
	// use sync.WaitGroup to prevent goroutine premature termination when main exits
	go func() {
		file, err := os.OpenFile(st.Cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		defer wg.Done()
		encoder := json.NewEncoder(file)
		st.Encoder = encoder
		<-ctx.Done()
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
		defer s.mu.Unlock()
		URL, ok := s.DB[sURL]
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
		s.mu.Lock()
		defer s.mu.Unlock()
		_, ok := s.DB[sURL]
		if ok {
			dumpError <- "already exists in DB"
			return
		}
		s.DB[sURL] = URL
		err := s.addToFileDB(sURL, URL)
		if err != nil {
			dumpError <- "could not add to file DB"
			return
		}
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
		s.DB[entry.SURL] = entry.URL
	}
	return nil
}

// addToFileDB adds one sURL:URL key-value pair to a file DB.
func (s *Storage) addToFileDB(sURL, URL string) error {
	rowToEncode := modelstorage.URLStorageEntry{
		SURL: sURL,
		URL:  URL,
	}
	err := s.Encoder.Encode(rowToEncode)
	if err != nil {
		return err
	}
	log.Print("POST query was saved to DB")
	return nil
}
