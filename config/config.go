package config

import (
	"sync"
	"io/ioutil"
	"log"
	"encoding/json"
	"strings"
)

const defaultPort = 8080

var (
	configMutex = sync.Mutex{}
	config      *Configuration
)

func Get() (*Configuration) {
	configMutex.Lock()
	defer configMutex.Unlock()

	if config == nil {
		config = &Configuration{}

		if bytes, err := ioutil.ReadFile("config.json"); err != nil {
			log.Fatal(err)
		} else {
			json.Unmarshal(bytes, config)
			config.Mantis.Host = strings.TrimRight(config.Mantis.Host, "/")
			if len(config.Secret) < 10 {
				log.Fatal("The configured secret is too small. Please set a new one on your config.json file.")
			}
			if config.Port == 0 {
				config.Port = defaultPort
			}
		}
	}

	return config
}

type Configuration struct {
	Port     int            `json:"port"`
	Secret   string         `json:"secret"`
	Database DatabaseConfig `json:"database"`
	Gitlab   GitlabConfig   `json:"gitlab"`
	Mantis   MantisConfig   `json:"mantis"`
}

type MantisConfig struct {
	Host     string `json:"host"`
	Token    string `json:"token"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type GitlabConfig struct {
	Token string `json:"token"`
}

type DatabaseConfig struct {
	Host         string `json:"host"`
	DatabaseName string `json:"database_name"`
	User         string `json:"user"`
	Password     string `json:"password"`
}
