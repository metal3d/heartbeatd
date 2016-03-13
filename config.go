package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Etcd  string          `yaml:"etcd"`
	Delay int             `yaml:"delay"`
	Keys  map[string]Test `yaml:"keys"`
}

func LoadConfig() *Config {
	conf := Config{"http://127.0.0.1:4001", 1, nil}
	content, err := ioutil.ReadFile("config.yml")
	if err == nil {
		yaml.Unmarshal(content, &conf)
	}
	return &conf
}
