package main

import (
	"log"
	"os"

	"github.com/upfluence/thrift-http-go/test/test"
	"github.com/upfluence/thrift/lib/go/thrift"
)

func main() {
	t, _ := thrift.NewTHttpPostClient("http://localhost:8080/foo")

	cl := test.NewTestClientFactory(t, thrift.NewTBinaryProtocolFactoryDefault())

	success := 0
	failed := 0
	for i := 0; i < 1000; i++ {
		_, err := cl.Add(int16(i), int16(i+1))

		if err != nil {
			failed++
		} else {
			success++
		}
	}

	log.Printf("Success: %d, Failure: %d", success, failed)

	if failed > 2 {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
