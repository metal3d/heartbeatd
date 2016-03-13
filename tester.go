package main

import (
	"log"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/client"
)

var checks = map[string]chan bool{}

// Test structure representing the yaml configuration "key".
type Test struct {
	Timeout       int
	Interval      time.Duration
	Test          string
	CommandFailed string `yaml:"command_on_fail"`
	CommandOK     string `yaml:"command_on_success"`
}

// Init will parse a key recursivally to initialize heartbeat.
func (test Test) Init(key string) {
	// initialize a watcher
	log.Println("Initialize watcher on", key)
	node, err := KAPI.Get(context.Background(), key, &client.GetOptions{
		Recursive: true,
	})
	if err == nil {
		if node.Node.Dir {
			for _, n := range node.Node.Nodes {
				test.Init(n.Key)
			}
			return
		}
		stop := make(chan bool)
		checklist <- Check{&test, node, stop}
	}
}

// Watch begins to watch a key and launches tests when a key moves.
func (test Test) Watch(key string) {
	log.Println("Watching key", key)
	test.Init(key)
	watcher := KAPI.Watcher(key, &client.WatcherOptions{
		Recursive: true,
	})

	for {
		node, err := watcher.Next(context.Background())
		if err != nil {
			log.Println(err)
			return
		}

		if stop, ok := checks[node.Node.Key]; ok {
			log.Println("CLEAN checker", node.Node.Key)
			stop <- true
			delete(checks, node.Node.Key)
		}

		if node.Action != "delete" {
			// prepare and start a check if the key was not deleted
			stop := make(chan bool)
			checklist <- Check{&test, node, stop}
		}
	}
}
