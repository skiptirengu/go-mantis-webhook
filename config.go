package main

import (
	"sync"
)

type Config struct {
	Token        string
	Repositories []repositoryConfig
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
	}

	return config
}
