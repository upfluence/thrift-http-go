package main

import (
	"fmt"
	"os"

	"github.com/upfluence/goutils/thrift"
	"github.com/upfluence/thrift-http-go/http_thrift"
	"github.com/upfluence/thrift-http-go/test/test"
)

type Handler struct{}

func (h *Handler) Add(x int16, y int16) (int16, error) {
	return x + y, nil
}

func main() {
	f, _ := os.Create("/tmp/test-http-thrift-go")
	f.WriteString(fmt.Sprintf("%d\n", os.Getpid()))
	f.Close()

	s, _ := http_thrift.NewTHTTPServer(":8080")
	thrift.NewServer(test.NewTestProcessor(&Handler{}), s).Start()
}
