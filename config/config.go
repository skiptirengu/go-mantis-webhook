package config

import (
	"sync"
	"io/ioutil"
	"encoding/json"
	"log"
)

var (
	config      *configs
	configMutex = sync.Mutex{}
)

func Get() (*configs) {
	configMutex.Lock()
	defer configMutex.Unlock()

	if config == nil {
		config = &configs{}

		if bytes, err := ioutil.ReadFile("config.json"); err != nil {
			log.Fatal(err)
		} else {
			json.Unmarshal(bytes, config)
		}
	}

	return config
}

type configs struct {
	Secret       string             `json:"secret"`
	Repositories []repositoryConfig `json:"repositories"`
	Database     databaseConfig     `json:"database"`
	Gitlab       gilabConfig        `json:"gitlab"`
	Mantis       mantisConfig       `json:"mantis"`
}

type mantisConfig struct {
	Host     string `json:"host"`
	Token    string `json:"token"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type gilabConfig struct {
	Token string `json:"token"`
}

type databaseConfig struct {
	Host         string `json:"host"`
	DatabaseName string `json:"database_name"`
	User         string `json:"user"`
	Password     string `json:"password"`
}

type repositoryConfig struct {
	Gitlab string
	Mantis string
}
