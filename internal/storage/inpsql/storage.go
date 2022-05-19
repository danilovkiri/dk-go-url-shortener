package inpsql

import (
	"context"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/modelurl"
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

// InitStorage initializes a Storage object, sets its attributes and starts a listener for persistStorage.
func InitStorage(ctx context.Context, wg *sync.WaitGroup, cfg *config.StorageConfig) (*Storage, error) {
	db, err := sqlx.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	st := Storage{
		Cfg: cfg,
		DB:  db,
	}
	err = st.createTable()
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
	return URL, nil
}

// RetrieveByUserID returns a slice of URL:sURL pairs defined as modelurl.FullURL for one particular user ID.
func (s *Storage) RetrieveByUserID(ctx context.Context, userID string) (URLs []modelurl.FullURL, err error) {
	return URLs, nil
}

// Dump stores a pair of sURL and URL as a key-value pair in a map.
func (s *Storage) Dump(ctx context.Context, URL string, sURL string, userID string) error {
	return nil
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
func (s *Storage) createTable() error {
	query := `CREATE TABLE IF NOT EXISTS urls (
		id bigserial not null,
		user_id uuid not null,
		url text not null,
		short_url text not null unique 
	);`
	_, err := s.DB.Exec(query)
	return err
}
