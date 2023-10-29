package storage

import (
	"bufio"
	"encoding/json"
	"os"
)
type URLData struct {
	UUID string `json:"uuid"`
	ShortURL string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type URLStorageSaver struct{
	file *os.File
	encoder *json.Encoder
}

var URLs map[string]string

var StorageSaver* URLStorageSaver

func InitURLStorage(filename string) error{
	URLs = make(map[string]string)
	if filename == "" {
		// значит опция сохранения в файл отключена
		return nil
	}
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
			return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan(){
		var urlData URLData
		err := json.Unmarshal(scanner.Bytes(), &urlData)
		if err != nil{
			return err
		}
		URLs[urlData.ShortURL] = urlData.OriginalURL
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}


func InitStorageSaver(filename string) error {

	if filename == "" {
		// значит опция сохранения в файл отключена
		StorageSaver = nil
		return nil
	}
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE| os.O_APPEND, 0666)
	if err != nil {
			return err
	}
	StorageSaver = &URLStorageSaver{
		file: file,
		encoder: json.NewEncoder(file),
	}
	return nil
}


func (storageSaver *URLStorageSaver) Save(urlData *URLData) error{
	return storageSaver.encoder.Encode(urlData)
}

func (storageSaver *URLStorageSaver) Close(){
	storageSaver.file.Close()
}