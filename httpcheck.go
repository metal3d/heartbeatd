package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/coreos/etcd/client"
)

// makeHttpTest tries to get a http url.
func makeHttpTest(test *Test, node *client.Response) error {

	log.Println("Make a HTTP test")
	resp, err := http.Get(node.Node.Value)
	if err != nil || resp.StatusCode > 399 {
		log.Println("HTTP failed, remove key")
		return errors.New("Failed")
	}
	defer resp.Body.Close()
	return nil
}
