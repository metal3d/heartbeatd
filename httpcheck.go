package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/coreos/etcd/client"
)

// makeHttpTest tries to get a http url.
func makeHttpTest(test *Test, node *client.Response) error {

	value, err := test.parseValue(node)
	if err != nil {
		return err
	}
	log.Println("Make a HTTP test", value)
	resp, err := http.Get(value)
	if err != nil || resp.StatusCode > 399 {
		return errors.New("Failed")
	}
	defer resp.Body.Close()
	return nil
}
