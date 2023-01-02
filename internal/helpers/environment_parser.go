// Package main provides ...
package helpers

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Feeds             []string `yaml:"Feeds"`
	ItemSince         float64  `yaml:"ItemSince"`
	ConcurrentWorkers int      `yaml:"ConcurrentWorkers"`
	CredentialPath    string   `yaml:"CredentialPath"`
}

var DEFAULT_CONFIG = &Config{
	Feeds: []string{
		"https://rsshub.app/theinitium/channel/latest/zh-hans",
	},
	ItemSince:         72.0,
	ConcurrentWorkers: 5,
}

func ParseConfig(fileName string) (*Config, error) {
	log.Printf("%v", fileName)

	if data, err := os.ReadFile(fileName); data != nil {
		log.Printf("%v", &data)
		t := Config{}

		if err := yaml.Unmarshal([]byte(data), &t); err != nil {
			log.Fatalf("error: %v", err)
			return nil, err
		}
		log.Printf("--- t:\n%v\n\n", t)
		return &t, nil
	} else {
		log.Fatalf("error: %v", err)
		return nil, err
	}
}
