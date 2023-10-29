package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
)

type serviceConfig struct {
	Host     string
	Port     int
	BaseAddr string
	Filename string
}

var ServiceConfig serviceConfig

func InitServiceConfig() error {
	var serviceAddr, baseAddr, filename string
	flag.StringVar(&serviceAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&baseAddr, "b", "http://localhost:8080", "base address of result shortened URL")
	flag.StringVar(&filename, "f", "", "filename of url storage")
	flag.Parse()
	if envServiceAddr := os.Getenv("SERVER_ADDRESS"); envServiceAddr != "" {
		serviceAddr = envServiceAddr
	}
	splittedAddr := strings.Split(serviceAddr, ":")
	if len(splittedAddr) != 2 {
		return errors.New("wrong format for flag -a")
	}
	ServiceConfig.Host = splittedAddr[0]
	var err error
	ServiceConfig.Port, err = strconv.Atoi(splittedAddr[1])
	if err != nil {
		return err
	}
	if envBaseAddr := os.Getenv("BASE_URL"); envBaseAddr != "" {
		baseAddr = envBaseAddr
	}
	ServiceConfig.BaseAddr = baseAddr
	if envFilename := os.Getenv("FILE_STORAGE_PATH"); envFilename != ""{
		filename = envFilename
	}
	ServiceConfig.Filename = filename
	return nil

}
