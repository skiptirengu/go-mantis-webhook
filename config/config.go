package config

import (
	"sync"
	"io/ioutil"
	"encoding/json"
	"log"
)

var (
	config      *Configs
	configMutex = sync.Mutex{}
)

func Get() (*Configs) {
	configMutex.Lock()
	defer configMutex.Unlock()

	if config == nil {
		config = &Configs{}

		if bytes, err := ioutil.ReadFile("config.json"); err != nil {
			log.Fatal(err)
		} else {
			json.Unmarshal(bytes, config)
		}
	}

	return config
}

type Configs struct {
	Port     int            `json:"port"`
	Secret   string         `json:"secret"`
	Database databaseConfig `json:"database"`
	Gitlab   gitlabConfig   `json:"gitlab"`
	Mantis   mantisConfig   `json:"mantis"`
}

type mantisConfig struct {
	Host     string `json:"host"`
	Token    string `json:"token"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type gitlabConfig struct {
	Token string `json:"token"`
}

type databaseConfig struct {
	Host         string `json:"host"`
	DatabaseName string `json:"database_name"`
	User         string `json:"user"`
	Password     string `json:"password"`
}
