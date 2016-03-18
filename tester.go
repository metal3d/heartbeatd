package main

import (
	"bytes"
	"log"
	"strings"
	"text/template"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/client"
)

// Test structure representing the yaml configuration "key".
type Test struct {
	Timeout       time.Duration
	Interval      time.Duration
	Test          string
	Value         string
	CommandFailed string `yaml:"on_fail"`
	CommandOK     string `yaml:"on_success"`
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
		done := make(chan bool)
		check := Check{&test, node, stop, done, false}
		go check.up()
		checkRegistry[node.Node.Key] = check
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

		if check, ok := checkRegistry[node.Node.Key]; ok {
			log.Println("CLEAN checker", node.Node.Key)
			check.stop <- true
			log.Println("DROP from registry", node.Node.Key)
			delete(checkRegistry, node.Node.Key)
		}

		if node.Action != "delete" {
			// prepare and start a check if the key was not deleted
			stop := make(chan bool)
			done := make(chan bool)
			check := Check{&test, node, stop, done, false}
			go check.up()
			checkRegistry[node.Node.Key] = check
		}
	}
}

// parseValue returns the parsed config test value template.
func (t Test) parseValue(node *client.Response) (string, error) {
	value := t.Value
	if len(strings.TrimSpace(value)) == 0 {
		value = "{{.Value}}"
	}
	tpl, err := template.New("Value").Parse(value)
	if err != nil {
		log.Println("Parse value error", err)
		return "", err
	}

	var b []byte
	buff := bytes.NewBuffer(b)
	err = tpl.Execute(buff, node.Node)
	return buff.String(), err
}
