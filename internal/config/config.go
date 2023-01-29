package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// Feeds is a list of RSS feeds to read from
	Feeds []string `yaml:"Feeds"`
	// ItemSince filter the feed based on the pub time in hours
	ItemSince float64 `yaml:"ItemSince"`
	// ConcurrentWorkers is the count of workers to start and process tss request
	ConcurrentWorkers int `yaml:"ConcurrentWorkers"`
	// CredentialPath is the path to the google credential
	CredentialPath string `yaml:"CredentialPath"`
	// MaxItemPerFeed is the max number of item per feed after applying the time
	// filter
	MaxItemPerFeed int `yaml:"MaxItemPerFeed"`
}

var DEFAULT_CONFIG = &Config{
	ItemSince:         24.0,
	ConcurrentWorkers: 5,
	MaxItemPerFeed:    10,
}

func NewConfig(fileName string) (*Config, error) {
	if data, err := os.ReadFile(fileName); data != nil {
		t := *DEFAULT_CONFIG

		if err := yaml.Unmarshal([]byte(data), &t); err != nil {
			return nil, err
		}

		if len(t.Feeds) == 0 || t.Feeds == nil {
			return nil, errors.New("no feed in the config file")
		}

		if t.CredentialPath == "" {
			return nil, errors.New("missing path to credential file in the config file")
		}

		return &t, nil
	} else {
		return nil, err
	}
}

func (t *Config) getFullCredentialPath() (fullPath string, err error) {
	if fullPath, err := filepath.Abs(t.CredentialPath); err == nil {
		return fullPath, nil
	} else {
		return "", err
	}
}
