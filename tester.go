package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/client"
)

type routine struct {
	key  string
	stop chan bool
}

var checks = map[string]routine{}

func newUUID() string {
	u := make([]byte, 16)
	_, err := rand.Read(u)
	if err != nil {
		return ""
	}

	u[8] = (u[8] | 0x80) & 0xBF // what does this do?
	u[6] = (u[6] | 0x40) & 0x4F // what does this do?

	return hex.EncodeToString(u)
}

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
		checks[newUUID()] = routine{node.Node.Key, stop}
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
		log.Println("Watcher", node.Node.Key, node.Node.Value, node.Action)

		if node.Action != "delete" {
			// add a checker for this key

			// remove old checkers because the key changed
			for i, r := range checks {
				if r.key == node.Node.Key {
					r.stop <- true
					delete(checks, i)
				}
			}

			// prepare and start a check
			stop := make(chan bool)
			checks[newUUID()] = routine{node.Node.Key, stop}
			checklist <- Check{&test, node, stop}
		} else {
			// node is deleted, so stop checks
			log.Println("DELETE", node)
			for i, r := range checks {
				if r.key == node.Node.Key {
					r.stop <- true
					delete(checks, i)
				}
			}
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
