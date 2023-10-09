package config

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

type serviceConfig struct {
	Host     string
	Port     int
	BaseAddr string
}

var ServiceConfig serviceConfig

func InitServiceConfig() error {
	var serviceAddr, baseAddr string
	flag.StringVar(&serviceAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&baseAddr, "b", "http://localhost:8000", "base address of result shortened URL")
	flag.Parse()
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
	ServiceConfig.BaseAddr = baseAddr
	return nil

}
