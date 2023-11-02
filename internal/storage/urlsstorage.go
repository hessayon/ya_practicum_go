package storage

import (
	"bufio"
	"encoding/json"
	"os"
)

type URLData struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type URLStorageSaver struct {
	file    *os.File
	encoder *json.Encoder
}

type URLStorage struct {
	Store    map[string]string
	filename string
	saver    *URLStorageSaver
}

var Storage *URLStorage

func NewURLStorage(filename string) (*URLStorage, error) {

	storage := &URLStorage{
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

func newStorageSaver(filename string) (*URLStorageSaver, error) {

	if filename == "" {
		return nil, nil
	}
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &URLStorageSaver{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (storage *URLStorage) Save(urlData *URLData) error {
	storage.Store[urlData.ShortURL] = urlData.OriginalURL
	if(storage.saver != nil){
		return storage.saver.encoder.Encode(urlData)
	}
	return nil
}


func (storage *URLStorage) Get(shortURL string) (string, bool) {
 	fullURL, found := storage.Store[shortURL]
	return fullURL, found
}


func (storage *URLStorage) Close() {
	if(storage.saver != nil){
		storage.saver.file.Close()
	}
}
