package main

import (
	"log"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
)

var KAPI client.KeysAPI

func main() {
	// load configuration
	cfg := LoadConfig()

	// set the number of parallel checks
	setParallel(cfg.Parallel)

	// create the etcd client
	c, err := client.New(client.Config{
		Endpoints:               strings.Split(cfg.Etcd, ","),
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	})

	if err != nil {
		log.Fatal(err, c)
	}

	// create the key api
	KAPI = client.NewKeysAPI(c)

	// and now.. watch !
	for key, test := range cfg.Keys {
		go test.Watch(key)
	}

	go func() {
		for {
			<-time.Tick(time.Second * 1)
			log.Println("Checks size:", checks)
		}
	}()

	<-make(chan bool)

}
