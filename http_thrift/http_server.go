package http_thrift

import (
	"errors"
	"net"
	"net/http"
	"sync"

	"github.com/upfluence/thrift/lib/go/thrift"
)

type THTTPServer struct {
	server     *http.Server
	listener   *net.Listener
	deliveries chan *THTTPRequest

	mu          sync.RWMutex
	interrupted bool
}

func NewTHTTPServerFromMux(
	mux *http.ServeMux,
	pattern string,
) (*THTTPServer, error) {
	server := &THTTPServer{deliveries: make(chan *THTTPRequest)}
	mux.Handle(pattern, &HTTPHandler{server})

	return server, nil
}

func NewTHTTPServer(listenAddr string) (*THTTPServer, error) {
	l, err := net.Listen("tcp", listenAddr)

	if err != nil {
		return nil, err
	}

	thriftServer := &THTTPServer{
		deliveries: make(chan *THTTPRequest),
		listener:   &l,
	}

	thriftServer.server = &http.Server{Handler: &HTTPHandler{thriftServer}}

	return thriftServer, nil
}

func (p *THTTPServer) Listen() error {
	if p.server != nil && p.listener != nil {
		go p.server.Serve(*p.listener)
	}

	return nil
}

func (s *THTTPServer) Accept() (thrift.TTransport, error) {
	s.mu.RLock()
	interrupted := s.interrupted
	s.mu.RUnlock()

	if interrupted {
		return nil, errors.New("Transport Interrupted")
	}

	return <-s.deliveries, nil
}

func (p *THTTPServer) Close() error {
	if p.listener != nil {
		return (*p.listener).Close()
	}

	return nil
}

func (p *THTTPServer) Interrupt() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.interrupted = true
	return nil
}
