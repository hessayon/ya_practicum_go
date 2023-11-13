package storage

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
)

type URLStorage interface {
	Save(urlData *URLData) (err error)
	SaveBatch(urlsBatch []*URLData) (err error) 
	Get(shortURL string) (value string, ok bool)
	Close()

}

type URLData struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
//--------------------------------------------------------------------

type URLStorageFileSaver struct {
	file    *os.File
	encoder *json.Encoder
}

type LocalURLStorage struct {
	Store    map[string]string
	filename string
	saver    *URLStorageFileSaver
}
//--------------------------------------------------------------------

type URLDBStorage struct {
	DB *sql.DB
}
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
	storage.Store[urlData.ShortURL] = urlData.OriginalURL
	if(storage.saver != nil){
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


func (storage *LocalURLStorage) Get(shortURL string) (string, bool) {
 	fullURL, found := storage.Store[shortURL]
	return fullURL, found
}


func (storage *LocalURLStorage) Close() {
	if(storage.saver != nil){
		storage.saver.file.Close()
	}
}
//--------------------------------------------------------------------


func (storage *URLDBStorage) createTable() error {
	query := `
	CREATE TABLE urls (
		short_url varchar NOT NULL,
		full_url varchar NOT NULL,
		CONSTRAINT urls_pk PRIMARY KEY (short_url)
	);`
	_, err := storage.DB.ExecContext(context.Background(), query)
	return err
}


func (storage *URLDBStorage) Save(urlData *URLData) error {
	query := "INSERT INTO urls VALUES ($1, $2);"
	_, err := storage.DB.ExecContext(context.Background(), query, urlData.ShortURL, urlData.OriginalURL)
	if err != nil {
		err = storage.createTable()
		
		if err != nil {
			return err
		}
		_, err := storage.DB.ExecContext(context.Background(), query, urlData.ShortURL, urlData.OriginalURL)
		return err
	}
	return nil
}


func (storage *URLDBStorage) SaveBatch(urlsBatch []*URLData) error {
	query := "INSERT INTO urls VALUES ($1, $2);"
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
	for i, data := range urlsBatch {
		_, err := stmt.ExecContext(ctx, data.ShortURL, data.OriginalURL)
		if err != nil {
			// если первая попытка сохранить, то пробуем создать таблицу
			if i == 0 {
				err = storage.createTable()
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


func (storage *URLDBStorage) Get(shortURL string) (string, bool) {
	query := "SELECT full_url FROM urls WHERE short_url = $1 LIMIT 1"
	row := storage.DB.QueryRowContext(context.Background(), query, shortURL)
	var fullURL string
	err := row.Scan(&fullURL)
	if err != nil{
		log.Printf("Error in Scan: %s", err.Error())
		return "", false
	}
	return fullURL, true
}


func (storage *URLDBStorage) Close() {
	storage.DB.Close()
}

func NewDBURLStorage(dsn string) (URLStorage, error) {
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



func NewURLStorage(filename string) (URLStorage, error) {

	storage := &LocalURLStorage{
		Store:    make(map[string]string),
		filename: filename,
		saver: nil,
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
		storage.Store[urlData.ShortURL] = urlData.OriginalURL
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
