package inpsql

import (
	"context"
	"database/sql"
	"errors"
	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
	"github.com/danilovkiri/dk_go_url_shortener/internal/service/modelurl"
	storageErrors "github.com/danilovkiri/dk_go_url_shortener/internal/storage/errors"
	"github.com/danilovkiri/dk_go_url_shortener/internal/storage/modelstorage"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/lib/pq"
	"log"
	"sync"
)

// Storage struct defines data structure handling and provides support for adding new implementations.
type Storage struct {
	mu  sync.Mutex
	Cfg *config.StorageConfig
	DB  *sql.DB
}

// InitStorage initializes a Storage object and sets its attributes.
func InitStorage(ctx context.Context, wg *sync.WaitGroup, cfg *config.StorageConfig) (*Storage, error) {
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
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

// Retrieve returns a URL corresponding to sURL.
func (s *Storage) Retrieve(ctx context.Context, sURL string) (URL string, err error) {
	// prepare query statement
	selectStmt, err := s.DB.PrepareContext(ctx, "SELECT * FROM urls WHERE short_url = $1")
	if err != nil {
		return "", &storageErrors.StatementPSQLError{Err: err}
	}
	defer selectStmt.Close()

	// create channels for listening to the go routine result
	retrieveDone := make(chan string)
	retrieveError := make(chan error)
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		var queryOutput modelstorage.URLPostgresEntry
		err := selectStmt.QueryRowContext(ctx, sURL).Scan(&queryOutput.ID, &queryOutput.UserID, &queryOutput.URL, &queryOutput.SURL, &queryOutput.IsDeleted)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				retrieveError <- &storageErrors.NotFoundError{Err: err, SURL: sURL}
				return
			default:
				retrieveError <- err
				return
			}
		}
		if queryOutput.IsDeleted {
			retrieveError <- &storageErrors.DeletedError{Err: err, SURL: sURL}
			return
		}
		retrieveDone <- queryOutput.URL
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
	// prepare query statement
	selectStmt, err := s.DB.PrepareContext(ctx, "SELECT * FROM urls WHERE user_id = $1 AND is_deleted = false")
	if err != nil {
		return nil, &storageErrors.StatementPSQLError{Err: err}
	}
	defer selectStmt.Close()

	// create channels for listening to the go routine result
	retrieveDone := make(chan []modelurl.FullURL)
	retrieveError := make(chan error)
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		rows, err := selectStmt.QueryContext(ctx, userID)
		if err != nil {
			retrieveError <- &storageErrors.ExecutionPSQLError{Err: err}
			return
		}
		defer rows.Close()

		// extract DB row data into corresponding go structure
		var queryOutput []modelstorage.URLPostgresEntry
		for rows.Next() {
			var queryOutputRow modelstorage.URLPostgresEntry
			err = rows.Scan(&queryOutputRow.ID, &queryOutputRow.UserID, &queryOutputRow.URL, &queryOutputRow.SURL, &queryOutputRow.IsDeleted)
			if err != nil {
				retrieveError <- &storageErrors.ScanningPSQLError{Err: err}
				return
			}
			queryOutput = append(queryOutput, queryOutputRow)
		}
		err = rows.Err()
		if err != nil {
			retrieveError <- &storageErrors.ScanningPSQLError{Err: err}
		}
		// extract go structure data into necessary output structure
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
		return nil, &storageErrors.ContextTimeoutExceededError{Err: ctx.Err()}
	case rtrvError := <-retrieveError:
		log.Println("Retrieving URLs by user ID:", rtrvError.Error())
		return nil, rtrvError
	case URLs := <-retrieveDone:
		log.Println("Retrieving URLs by user ID:", URLs)
		return URLs, nil
	}
}

// Dump stores a pair of sURL and URL as a key-value pair in a map.
func (s *Storage) Dump(ctx context.Context, URL string, sURL string, userID string) error {
	// prepare INSERT statement
	dumpStmt, err := s.DB.PrepareContext(ctx, "INSERT INTO urls (user_id, url, short_url) VALUES ($1, $2, $3)")
	if err != nil {
		return &storageErrors.StatementPSQLError{Err: err}
	}
	defer dumpStmt.Close()
	// prepare SELECT statement
	selectStmt, err := s.DB.PrepareContext(ctx, "SELECT short_url FROM urls WHERE url = $1")
	if err != nil {
		return &storageErrors.StatementPSQLError{Err: err}
	}
	defer selectStmt.Close()

	// create channels for listening to the go routine result
	dumpDone := make(chan bool)
	dumpError := make(chan error)
	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		_, err := dumpStmt.ExecContext(ctx, userID, URL, sURL)
		if err != nil {
			if err, ok := err.(*pgconn.PgError); ok && err.Code == pgerrcode.UniqueViolation {
				// retrieve already existing sURL for violating unique constraint URL
				var validsURL string
				err := selectStmt.QueryRowContext(ctx, URL).Scan(&validsURL)
				if err != nil {
					dumpError <- &storageErrors.ExecutionPSQLError{Err: err}
					return
				}
				dumpError <- &storageErrors.AlreadyExistsError{Err: err, URL: URL, ValidSURL: validsURL}
				return
			}
			dumpError <- &storageErrors.ExecutionPSQLError{Err: err}
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

// DeleteBatch assigns a deletion flag for DB entries.
func (s *Storage) DeleteBatch(ctx context.Context, sURLs []string, userID string) error {
	// prepare DELETE statement
	deleteStmt, err := s.DB.PrepareContext(ctx, "UPDATE urls SET is_deleted = true WHERE user_id = $1 AND short_url = ANY($2)")
	if err != nil {
		return &storageErrors.StatementPSQLError{Err: err}
	}
	defer deleteStmt.Close()
	//begin transaction
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return &storageErrors.ExecutionPSQLError{Err: err}
	}
	defer tx.Rollback()
	txDeleteStmt := tx.StmtContext(ctx, deleteStmt)
	// create channels for listening to the go routine result
	deleteDone := make(chan bool)
	deleteError := make(chan error)
	go func() {
		_, err = txDeleteStmt.ExecContext(
			ctx,
			userID,
			pq.Array(sURLs),
		)
		if err != nil {
			deleteError <- &storageErrors.ExecutionPSQLError{Err: err}
		}
		deleteDone <- true
	}()

	// wait for the first channel to retrieve a value
	select {
	case <-ctx.Done():
		log.Println("Deleting URL:", ctx.Err())
		return &storageErrors.ContextTimeoutExceededError{Err: ctx.Err()}
	case dltError := <-deleteError:
		log.Println("Deleting URL:", dltError.Error())
		return dltError
	case <-deleteDone:
		log.Println("Deleting URL:", sURLs)
		return tx.Commit()
	}
}

// PingDB performs DB ping.
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
		url text not null unique,
		short_url text not null,
		is_deleted boolean not null DEFAULT false 
	);`
	_, err := s.DB.ExecContext(ctx, query)
	return err
}
