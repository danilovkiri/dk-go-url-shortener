package inpsql_sqlx

import (
	"context"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/modelurl"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/errors"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/modelstorage"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"log"
	"sync"
)

// Storage struct defines data structure handling and provides support for adding new implementations.
type Storage struct {
	mu  sync.Mutex
	Cfg *config.StorageConfig
	DB  *sqlx.DB
}

// InitStorage initializes a Storage object and sets its attributes.
func InitStorage(ctx context.Context, wg *sync.WaitGroup, cfg *config.StorageConfig) (*Storage, error) {
	db, err := sqlx.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	st := Storage{
		Cfg: cfg,
		DB:  db,
	}
	err = st.createTable(ctx)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer wg.Done()
		<-ctx.Done()
		err := st.DB.Close()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("PSQL DB connection closed successfully")
	}()
	return &st, nil
}

// Retrieve returns a URL as a value of a map based on the given sURL as a key of a map.
func (s *Storage) Retrieve(ctx context.Context, sURL string) (URL string, err error) {
	query := "SELECT url FROM urls WHERE short_url = $1"
	// create channels for listening to the go routine result
	retrieveDone := make(chan string)
	retrieveError := make(chan string)
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		// use GetContext due to one variable usage as an output
		err = s.DB.GetContext(ctx, &URL, query, sURL)
		if err != nil {
			retrieveError <- "PSQL error"
			return
		}
		if len(URL) == 0 {
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
		if errString == "PSQL error" {
			return "", errors.StoragePSQLError{}
		}
		return "", errors.StorageNotFoundError{ID: sURL}
	case URL := <-retrieveDone:
		log.Println("Retrieving URL:", sURL, "as", URL)
		return URL, nil
	}
}

// RetrieveByUserID returns a slice of URL:sURL pairs defined as modelurl.FullURL for one particular user ID.
func (s *Storage) RetrieveByUserID(ctx context.Context, userID string) (URLs []modelurl.FullURL, err error) {
	query := "SELECT * FROM urls WHERE user_id = $1"
	// create channels for listening to the go routine result
	retrieveDone := make(chan []modelurl.FullURL)
	retrieveError := make(chan string)
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		var queryOutput []modelstorage.URLPostgresEntry
		// use SelectContext due to struct usage and slices
		err := s.DB.SelectContext(ctx, &queryOutput, query, userID)
		if err != nil {
			retrieveError <- "PSQL error"
			return
		}
		var URLs []modelurl.FullURL
		for _, entry := range queryOutput {
			fullURL := modelurl.FullURL{
				URL:  entry.URL,
				SURL: entry.SURL,
			}
			URLs = append(URLs, fullURL)
		}
		retrieveDone <- URLs
	}()

	// wait for the first channel to retrieve a value
	select {
	case <-ctx.Done():
		log.Println("Retrieving URLs by user ID:", ctx.Err())
		return nil, errors.ContextTimeoutExceededError{}
	case errString := <-retrieveError:
		log.Println("Retrieving URLs by user ID:", errString)
		return nil, errors.StoragePSQLError{}
	case URLs := <-retrieveDone:
		log.Println("Retrieving URLs by user ID:", URLs)
		return URLs, nil
	}
}

// Dump stores a pair of sURL and URL as a key-value pair in a map.
func (s *Storage) Dump(ctx context.Context, URL string, sURL string, userID string) error {
	query := "INSERT INTO urls (user_id, url, short_url) VALUES ($1, $2, $3)"
	// create channels for listening to the go routine result
	dumpDone := make(chan bool)
	dumpError := make(chan string)
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		_, err := s.DB.ExecContext(ctx, query, userID, URL, sURL)
		if err != nil {
			dumpError <- "PSQL error"
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
		return errors.StoragePSQLError{}
	case <-dumpDone:
		log.Println("Dumping URL:", sURL, "as", URL)
		return nil
	}
}

// PingDB is a mock for PSQL DB pinger for infile DB handling.
func (s *Storage) PingDB() error {
	return s.DB.Ping()
}

// CloseDB is a mock for PSQL DB closer for infile DB handling.
func (s *Storage) CloseDB() error {
	return s.DB.Close()
}

// createTable creates a table for PSQL DB storage if not exist.
func (s *Storage) createTable(ctx context.Context) error {
	// store user_id as text since we store encoded tokens
	query := `CREATE TABLE IF NOT EXISTS urls (
		id bigserial not null,
		user_id text not null,
		url text not null,
		short_url text not null unique 
	);`
	_, err := s.DB.ExecContext(ctx, query)
	return err
}
