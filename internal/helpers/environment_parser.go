// Package main provides ...
package helpers

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Feeds             []string `yaml:"Feeds"`
	ItemSince         float64  `yaml:"ItemSince"`
	ConcurrentWorkers int      `yaml:"ConcurrentWorkers"`
	CredentialPath    string   `yaml:"CredentialPath"`
	MaxItemPerFeed    int      `yaml:"MaxItemPerFeed"`
}

var DEFAULT_CONFIG = &Config{
	Feeds: []string{
		"https://rsshub.app/theinitium/channel/latest/zh-hans",
	},
	ItemSince:         72.0,
	ConcurrentWorkers: 5,
}

func InitConfig(fileName string) (*Config, error) {
	if data, err := os.ReadFile(fileName); data != nil {
		t := Config{}

		if err := yaml.Unmarshal([]byte(data), &t); err != nil {
			log.Fatalf("error: %v", err)
			return nil, err
		}

		t.CredentialPath, _ = filepath.Abs(t.CredentialPath)
		return &t, nil
	} else {
		log.Fatalf("fail to read config file: %v", err)
		return nil, err
	}
}
