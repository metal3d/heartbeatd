package main

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Etcd     string          `yaml:"etcd"`
	Delay    int             `yaml:"delay"`
	Parallel int             `yaml:"parallel"`
	Keys     map[string]Test `yaml:"keys"`
}

var (
	currentUser, _ = user.Current()
	CONFIGPATH     = []string{
		"./config.yml",
		filepath.Join(currentUser.HomeDir, ".config/heartbeatd/config.yml"),
		"/etc/heartbeatd/config.yml",
	}
)

// LoadConfig read configuration file and return a configuration objet.
func LoadConfig() *Config {
	conf := Config{"http://127.0.0.1:4001", 1, runtime.NumCPU(), nil}
	for _, p := range CONFIGPATH {
		if _, err := os.Stat(p); err != nil {
			content, err := ioutil.ReadFile("config.yml")
			if err == nil {
				yaml.Unmarshal(content, &conf)
				break
			}
		}
	}
	return &conf
}
