package main

import (
	"sync"
	"io/ioutil"
	"encoding/json"
	"log"
)

type Config struct {
	Token        string             `json:"token"`
	Repositories []repositoryConfig `json:"repositories"`
	Database     databaseConfig     `json:"database"`
}

type databaseConfig struct {
	Host         string `json:"host"`
	Port         string `json:"port"`
	DatabaseName string `json:"database_name"`
	User         string `json:"user"`
	Password     string `json:"password"`
}

type repositoryConfig struct {
	Gitlab string
	Mantis string
}

var (
	config      *Config
	configMutex = sync.Mutex{}
)

func GetConfig() (*Config) {
	configMutex.Lock()
	defer configMutex.Unlock()

	if config == nil {
		config = &Config{}

		if bytes, err := ioutil.ReadFile("config.json"); err != nil {
			log.Fatal(err)
		} else {
			json.Unmarshal(bytes, config)
		}
	}

	return config
}
