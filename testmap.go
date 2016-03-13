package main

import "github.com/coreos/etcd/client"

// testfunc is a function that can make a test.
type testfunc func(*Test, *client.Response) error

// TESTS stores corresponding "test" (from config.yaml)
// to a testfunc.
var TESTS = map[string]testfunc{
	"connect": makeSocketTest,
	"http":    makeHttpTest,
}
