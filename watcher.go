package main

import (
	"log"
	"time"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// Watcher represents a watched test.
type Watcher struct {
	test Test
}

var checks = make(map[string]chan bool)

func (w *Watcher) Watch(key string) {
	watcher := KAPI.Watcher(key, &client.WatcherOptions{
		Recursive: true,
	})
	for {
		node, err := watcher.Next(context.Background())
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Watcher", node.Node.Key, node.Node.Value, node.Action)

		if node.Action != "delete" {
			// add a checker for this key
			if quit, exists := checks[node.Node.Key]; exists {
				// a watcher exists, maybe the value changed, so we remove
				// the current check by stopping its goroutine
				// and we will create another
				quit <- true
			}

			// prepare and start a check
			stop := make(chan bool)
			checks[node.Node.Key] = stop
			go checkup(w, node, stop)
		} else {
			// node is deleted, so stop checks
			if stop, exists := checks[node.Node.Key]; exists {
				stop <- true
				delete(checks, node.Node.Key)
			}
		}
	}
}

// checkup makes periodical test
func checkup(w *Watcher, node *client.Response, stop chan bool) {
	for {
		select {
		case <-stop:
			log.Println("Stop Checks")
			return
		case <-time.Tick(time.Second * 1):
			MakeTest(w, node)
		}
	}
}
