package main

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/coreos/etcd/client"
)

type Test struct {
	Timeout       int
	Test          string
	Actions       []string
	CommandFailed string `yaml:"command_on_fail"`
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
func MakeTest(w *Watcher, node *client.Response) {
	t := &w.test
	var err error

	switch t.Test {
	case "http":
		err = makeHttpTest(t, node)
	case "connect":
		makeSocketTest(t, node)
	}

	if err != nil {
		//w.kapi.Delete(context.Background(), node.Node.Key, nil)
		if t.CommandFailed == "" {
			log.Println("No command for failed state specified")
			return
		}
		// create a parsed command
		cmd, err := getCommand(t.CommandFailed, node)
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
}

func makeHttpTest(t *Test, node *client.Response) error {

	log.Println("Make a HTTP test")
	resp, err := http.Get(node.Node.Value)
	if err != nil || resp.StatusCode > 399 {
		log.Println("HTTP failed, remove key")
		return errors.New("Failed")
	}

	return nil
}

func makeSocketTest(t *Test, node *client.Response) {
	log.Println("Make a Socket test")
}
