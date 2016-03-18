package main

import (
	"bytes"
	"html/template"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/coreos/etcd/client"
)

// Check registry to be able to retrieve
// "stop" channels, keys and so on...
var checkRegistry = map[string]Check{}

// Channel that recieves "Checks" to launch. This channel is
// made to scale checks and prevent overhead. See setParallel().
var checkLauncher chan *Check

// Check contains test for a key, the response from etcd and a stop channel
type Check struct {
	test       *Test
	node       *client.Response
	stop       chan bool
	done       chan bool
	inprogress bool
}

// up makes periodical test.
func (c *Check) up() {
	for {
		select {
		case <-c.stop:
			// We ask to stop test, try to wait a current test then quit.
			log.Println("STOPPING goroutine", c.node.Node.Key)
			if c.inprogress {
				log.Println("WAITING...")
				select {
				case <-time.Tick(c.test.Timeout):
					log.Println("... Timeout")
				case <-c.done:
					log.Println("... Done")
				}
			}
			log.Println("STOPPED goroutine", c.node.Node.Key)
			return
		case <-time.Tick(c.test.Interval):
			// send test to the stack. Do it when runtime
			// is not too busy.
			c.inprogress = true
			checkLauncher <- c
			<-c.done
		}
	}
}

// MakeTest launches test.
func (c *Check) makeTest() {
	var err error

	defer func() {
		c.inprogress = false
		c.done <- true
	}()

	// find a test and execute it
	if fnc, ok := TESTS[c.test.Test]; ok {
		err = fnc(c.test, c.node)
	} else {
		log.Println(c.test.Test, "is not a known test")
		return
	}

	if err != nil {
		//w.kapi.Delete(context.Background(), node.Node.Key, nil)
		if c.test.CommandFailed == "" {
			log.Println("No command for failed state specified")
			return
		}
		execCommand(c.test.CommandFailed, c.node)
	} else {
		if c.test.CommandOK == "" {
			return
		}
		execCommand(c.test.CommandOK, c.node)
	}
}

// parseCommand returns a parsed command from configuration.
func parseCommand(cmd string, node *client.Response) (*exec.Cmd, error) {

	tpl, err := template.New("Cmd").Parse(cmd)
	if err != nil {
		return nil, err
	}

	var b []byte
	buff := bytes.NewBuffer(b)
	if err := tpl.Execute(buff, node.Node); err != nil {
		return nil, err
	}

	args := []string{"-c", buff.String()}

	return exec.Command("sh", args...), nil
}

func execCommand(command string, node *client.Response) {
	// create a parsed command
	cmd, err := parseCommand(command, node)
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

// initialize the parallels routines
func setParallel(size int) {
	//checklist = make(chan Check, size)
	runtime.GOMAXPROCS(size)
	checkLauncher = make(chan *Check, size)
	go func() {
		// as soon as we recieve a check, run it !
		for c := range checkLauncher {
			c.makeTest()
		}
	}()
}
