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

	KAPI = client.NewKeysAPI(c)

	for key, test := range cfg.Keys {
		go test.Watch(key)
	}

	<-make(chan bool)

}
