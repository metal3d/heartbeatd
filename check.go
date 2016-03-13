package main

import (
	"log"
	"runtime"
	"time"

	"github.com/coreos/etcd/client"
)

var checklist chan Check

type Check struct {
	test *Test
	node *client.Response
	stop chan bool
}

// up makes periodical test
func (c Check) up() {
	for {
		select {
		case <-c.stop:
			log.Println("STOP check goroutine", c.node)
			return
		case <-time.Tick(c.test.Interval * time.Second):
			c.makeTest()
		}
	}
}

// MakeTest launches test.
func (c Check) makeTest() {
	var err error

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

// initialize the parallels routines
func setParallel(size int) {
	checklist = make(chan Check, size)
	runtime.GOMAXPROCS(size)
	log.Println("Launching", size, "check goroutines")
	for i := 0; i < size; i++ {
		go func() {
			for check := range checklist {
				check.up()
			}
		}()
	}
}
