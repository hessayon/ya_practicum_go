package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
)

type ServiceConfig struct {
	Host     string
	Port     int
	BaseAddr string
	Filename string
	DBDsn    string
}

var Config *ServiceConfig

func NewServiceConfig() (*ServiceConfig, error) {

	var serviceAddr, baseAddr, filename, dbDSN string
	flag.StringVar(&serviceAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&baseAddr, "b", "http://localhost:8080", "base address of result shortened URL")
	flag.StringVar(&filename, "f", "", "filename of url storage")
	flag.StringVar(&dbDSN, "d", "", "database connection string")
	flag.Parse()
	if envServiceAddr := os.Getenv("SERVER_ADDRESS"); envServiceAddr != "" {
		serviceAddr = envServiceAddr
	}
	splittedAddr := strings.Split(serviceAddr, ":")
	if len(splittedAddr) != 2 {
		return nil, errors.New("wrong format for flag -a")
	}
	host := splittedAddr[0]
	port, err := strconv.Atoi(splittedAddr[1])
	if err != nil {
		return nil, err
	}

	if envBaseAddr := os.Getenv("BASE_URL"); envBaseAddr != "" {
		baseAddr = envBaseAddr
	}

	if envFilename := os.Getenv("FILE_STORAGE_PATH"); envFilename != "" {
		filename = envFilename
	}

	if envDBDSN := os.Getenv("DATABASE_DSN"); envDBDSN != "" {
		dbDSN = envDBDSN
	}

	return &ServiceConfig{
		Host:     host,
		Port:     port,
		BaseAddr: baseAddr,
		Filename: filename,
		DBDsn:    dbDSN,
	}, nil

}

func NewDefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		Host:     "",
		Port:     8080,
		BaseAddr: "http://localhost:8080",
		Filename: "",
	}
}
