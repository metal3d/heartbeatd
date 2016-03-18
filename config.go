package main

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	currentUser, _ = user.Current()
	CONFIGPATH     = []string{
		"./config.yml",
		filepath.Join(currentUser.HomeDir, ".config/heartbeatd/config.yml"),
		"/etc/heartbeatd/config.yml",
	}
)

type Config struct {
	Etcd     string
	Interval time.Duration
	Parallel int
	Keys     map[string]*KeyConf
}

// UnmarshalYAML fixes some initial variables.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var cfg struct {
		Etcd     string              `yaml:"etcd"`
		Interval time.Duration       `yaml:"interval"`
		Parallel int                 `yaml:"parallel"`
		Keys     map[string]*KeyConf `yaml:"keys"`
	}

	if err := unmarshal(&cfg); err != nil {
		return err
	}

	// fixup times to "seconds"
	if cfg.Interval < 1 {
		cfg.Interval = 1 * time.Second
	} else {
		cfg.Interval = cfg.Interval * time.Second
	}

	if cfg.Parallel < 2 {
		cfg.Parallel = runtime.NumCPU() + 1
	}

	for _, v := range cfg.Keys {
		if v.Interval < 1 {
			v.Interval = cfg.Interval
		} else {
			v.Interval = v.Interval * time.Second
		}
		if v.Timeout < 1 {
			v.Timeout = cfg.Interval
		} else {
			v.Timeout = v.Timeout * time.Second
		}
	}

	c.Etcd = cfg.Etcd
	c.Parallel = cfg.Parallel
	c.Interval = cfg.Interval
	c.Keys = cfg.Keys
	return nil
}

// LoadConfig read configuration file and return a configuration objet.
func LoadConfig() *Config {
	conf := Config{}
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
