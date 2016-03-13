package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/client"
)

var checks = make(map[string]chan bool)

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
		checks[node.Node.Key] = stop
		go checkup(&test, node, stop)
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
			go checkup(&test, node, stop)
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
func checkup(test *Test, node *client.Response, stop chan bool) {
	for {
		select {
		case <-stop:
			return
		case <-time.Tick(test.Interval * time.Second):
			MakeTest(test, node)
		}
	}
}

// getCommand returns a parsed command from configuration.
func getCommand(cmd string, node *client.Response) (*exec.Cmd, error) {

	tpl, err := template.New("Cmd").Parse(cmd)
	if err != nil {
		return nil, err
	}

	var b []byte
	buff := bytes.NewBuffer(b)
	if err := tpl.Execute(buff, node.Node); err != nil {
		return nil, err
	}

	args := strings.Split(buff.String(), " ")

	if len(args) > 0 {
		return exec.Command(args[0], args[1:]...), nil
	}

	return exec.Command(args[0]), nil
}

// MakeTest launches test.
func MakeTest(test *Test, node *client.Response) {
	var err error

	// find a test and execute it
	if fnc, ok := TESTS[test.Test]; ok {
		fnc(test, node)
	} else {
		log.Println(test.Test, "is not a known test")
		return
	}

	if err != nil {
		//w.kapi.Delete(context.Background(), node.Node.Key, nil)
		if test.CommandFailed == "" {
			log.Println("No command for failed state specified")
			return
		}
		execCommand(test.CommandFailed, node)
	} else {
		if test.CommandOK == "" {
			return
		}
		execCommand(test.CommandOK, node)
	}
}

func execCommand(command string, node *client.Response) {
	// create a parsed command
	cmd, err := getCommand(command, node)
	if err != nil {
		log.Println(err)
		return
	}
	// Use stdin and stdout to see the result
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	// launch
	if err := cmd.Run(); err != nil {
		log.Println("Run", err)
	}
}
