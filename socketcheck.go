package main

import (
	"log"
	"net"

	"github.com/coreos/etcd/client"
)

// makeSocketTest tries to open a socket.
func makeSocketTest(test *Test, node *client.Response) error {
	log.Println("Make a Socket test")
	c, err := net.Dial("tcp", node.Node.Value)
	if err == nil {
		defer c.Close()
	}
	return err
}
