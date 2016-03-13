package main

import (
	"log"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
)

var KAPI client.KeysAPI

func main() {
	cfg := LoadConfig()

	c, err := client.New(client.Config{
		Endpoints:               strings.Split(cfg.Etcd, ","),
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	})

	if err != nil {
		log.Fatal(err, c)
	}

	watchers := make([]Watcher, 0)
	KAPI = client.NewKeysAPI(c)

	for key, conf := range cfg.Keys {
		w := Watcher{conf}
		watchers = append(watchers, w)
		go w.Watch(key)
	}

	<-make(chan bool)

}
