package storage

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/hessayon/ya_practicum_go/internal/config"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type URLStorage interface {
	Save(urlData *URLData) (err error)
	SaveBatch(urlsBatch []*URLData) (err error)
	GetOriginalURL(shortURL string) (value string, err error)
	GetShortURL(originalURL string) (value string, err error)
	GetURLsByUserID(userID string) (value []URLData, err error)
	DeleteURLs(urls...string) (err error)
	Close()
}

type URLData struct {
	UUID        string `json:"uuid,omitempty"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	DeletedFlag bool 		`json:"-"`
}

//--------------------------------------------------------------------

type URLStorageFileSaver struct {
	file    *os.File
	encoder *json.Encoder
}

type LocalURLStorage struct {
	ShortToOrig map[string]string
	OrigToShort map[string]string
	UUIDToData map[string][]URLData
	URLsToDelete []string
	filename    string
	saver       *URLStorageFileSaver
}

//--------------------------------------------------------------------

type URLDBStorage struct {
	DB *sql.DB
}

//--------------------------------------------------------------------

var ErrConflict = errors.New("data conflict")
var ErrAlreadyDeleted = errors.New("url is already deleted")
var ErrURLNotFound = errors.New("url is not found")

//--------------------------------------------------------------------

func newStorageSaver(filename string) (*URLStorageFileSaver, error) {

	if filename == "" {
		return nil, nil
	}
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &URLStorageFileSaver{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (storage *LocalURLStorage) Save(urlData *URLData) error {
	if _, ok := storage.OrigToShort[urlData.OriginalURL]; ok {
		return ErrConflict
	}
	storage.ShortToOrig[urlData.ShortURL] = urlData.OriginalURL
	storage.OrigToShort[urlData.OriginalURL] = urlData.ShortURL
	storage.UUIDToData[urlData.UUID] = append(storage.UUIDToData[urlData.UUID], *urlData)
	if storage.saver != nil {
		return storage.saver.encoder.Encode(urlData)
	}
	return nil
}

func (storage *LocalURLStorage) SaveBatch(urlsBatch []*URLData) error {
	for _, data := range urlsBatch {
		err := storage.Save(data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (storage *LocalURLStorage) GetOriginalURL(shortURL string) (string, error) {
	fullURL, found := storage.ShortToOrig[shortURL]
	if !found {
		return "", ErrURLNotFound
	}
	return fullURL, nil
}

func (storage *LocalURLStorage) GetShortURL(originalURL string) (string, error) {
	shortURL, found := storage.OrigToShort[originalURL]
	if !found {
		return "", ErrURLNotFound
	}
	return shortURL, nil
}

func (storage *LocalURLStorage) Close() {
	if storage.saver != nil {
		storage.saver.file.Close()
	}
}

func (storage *LocalURLStorage) GetURLsByUserID(userID string) ([]URLData, error) {
	resList, found := storage.UUIDToData[userID]
	if !found {
		errMsg := fmt.Sprintf("Not found data for %s", userID)
		return nil, errors.New(errMsg)
	}
	return resList, nil
}

func (storage *LocalURLStorage) DeleteURLs(urls...string) error {
	return nil
}

//--------------------------------------------------------------------

func (storage *URLDBStorage) createTable() error {
	query := `
	CREATE TABLE urls (
		short_url varchar NOT NULL,
		full_url varchar NOT NULL,
		uuid varchar NOT NULL,
		deleted boolean DEFAULT FALSE;
		CONSTRAINT urls_pk PRIMARY KEY (full_url)
	);`
	_, err := storage.DB.ExecContext(context.Background(), query)
	return err
}

func (storage *URLDBStorage) Save(urlData *URLData) error {
	query := "INSERT INTO urls VALUES ($1, $2, $3);"
	_, err := storage.DB.ExecContext(context.Background(), query, urlData.ShortURL, urlData.OriginalURL, urlData.UUID)
	if err != nil {
		var pgErr *pgconn.PgError
		// если не найдена такая таблица, то пробуем создать таблицу
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UndefinedTable {
			err = storage.createTable()
			if err != nil {
				return err
			}
			_, err := storage.DB.ExecContext(context.Background(), query, urlData.ShortURL, urlData.OriginalURL, urlData.UUID)
			return err
		} else if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			err = ErrConflict
			return err
		} else {
			log.Printf("unexpected error: %v", err)
			return err
		}
	}
	return nil
}

func (storage *URLDBStorage) SaveBatch(urlsBatch []*URLData) error {
	query := "INSERT INTO urls VALUES ($1, $2, $3);"
	tx, err := storage.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	ctx := context.Background()
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, data := range urlsBatch {
		_, err := stmt.ExecContext(ctx, data.ShortURL, data.OriginalURL, data.UUID)
		if err != nil {
			var pgErr *pgconn.PgError
			// если не найдена такая таблица, то пробуем создать таблицу
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UndefinedTable {
				err = storage.createTable()
				if err != nil {
					return err
				}
				_, err := stmt.ExecContext(ctx, data.ShortURL, data.OriginalURL, data.UUID)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	return tx.Commit()
}

func (storage *URLDBStorage) GetOriginalURL(shortURL string) (string, error) {
	query := "SELECT full_url, deleted FROM urls WHERE short_url = $1 LIMIT 1"
	row := storage.DB.QueryRowContext(context.Background(), query, shortURL)
	var fullURL string
	var isDeleted bool
	err := row.Scan(&fullURL, &isDeleted)
	if err != nil {
		log.Printf("Error in Scan: %s", err.Error())
		return "", ErrURLNotFound
	}
	if isDeleted {
		return "", ErrAlreadyDeleted
	}
	return fullURL, nil
}

func (storage *URLDBStorage) GetShortURL(originalURL string) (string, error) {
	query := "SELECT short_url FROM urls WHERE full_url = $1 LIMIT 1"
	row := storage.DB.QueryRowContext(context.Background(), query, originalURL)
	var shortURL string
	err := row.Scan(&shortURL)
	if err != nil {
		log.Printf("Error in Scan: %s", err.Error())
		return "", ErrURLNotFound
	}
	return shortURL, nil
}

func (storage *URLDBStorage) Close() {
	storage.DB.Close()
}

func (storage *URLDBStorage) GetURLsByUserID(userID string) ([]URLData, error) {
	resList := make([]URLData, 0)
	query := "SELECT short_url, full_url FROM urls WHERE uuid = $1"
	rows, err := storage.DB.QueryContext(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var urlData URLData
		err = rows.Scan(&urlData.ShortURL, &urlData.OriginalURL)
		if err != nil {
				return nil, err
		}
		resList = append(resList, urlData)	
	}	


	err = rows.Err()
	if err != nil {
			return nil, err
	}
	return resList, nil
}


func (storage *URLDBStorage) DeleteURLs(urls...string) error {
	return nil
}
//--------------------------------------------------------------------


func newDBURLStorage(dsn string) (*URLDBStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &URLDBStorage{
		DB: db,
	}, nil
}

func newLocalURLStorage(filename string) (*LocalURLStorage, error) {

	storage := &LocalURLStorage{
		ShortToOrig: make(map[string]string),
		OrigToShort: make(map[string]string),
		UUIDToData: make(map[string][]URLData),
		URLsToDelete: make([]string, 0),
		filename:    filename,
		saver:       nil,
	}

	if filename == "" {
		// значит опция сохранения в файл отключена
		return storage, nil
	}
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var urlData URLData
		err := json.Unmarshal(scanner.Bytes(), &urlData)
		if err != nil {
			return nil, err
		}
		storage.ShortToOrig[urlData.ShortURL] = urlData.OriginalURL
		storage.OrigToShort[urlData.OriginalURL] = urlData.ShortURL
		storage.UUIDToData[urlData.UUID] = append(storage.UUIDToData[urlData.UUID], urlData)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	storageSaver, err := newStorageSaver(filename)
	if err != nil {
		return nil, err
	}
	storage.saver = storageSaver
	return storage, nil
}


func NewURLStorage(cfg *config.ServiceConfig) (URLStorage, error) {
	if cfg.DBDsn != "" {
		return newDBURLStorage(cfg.DBDsn)
	} else {
		return newLocalURLStorage(cfg.Filename)
	}
}
