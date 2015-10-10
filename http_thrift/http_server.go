package http_thrift

import (
	"errors"
	"net/http"
	"sync"

	"github.com/upfluence/thrift/lib/go/thrift"
)

type THTTPRequest struct {
	w    *http.ResponseWriter
	r    *http.Request
	lock chan bool
}

func (d *THTTPRequest) Open() error {
	return nil
}

func (d *THTTPRequest) IsOpen() bool {
	return true
}

func (d *THTTPRequest) Close() error {
	d.lock <- true
	return nil
}

func (d *THTTPRequest) Read(buf []byte) (int, error) {
	return d.r.Body.Read(buf)
}

func (d *THTTPRequest) Write(buf []byte) (int, error) {
	return (*d.w).Write(buf)
}

func (d *THTTPRequest) Flush() error {
	d.lock <- true

	return nil
}

type THTTPServer struct {
	server     *http.Server
	deliveries chan *THTTPRequest

	mu          sync.RWMutex
	interrupted bool
}

type HTTPHandler struct {
	server *THTTPServer
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &THTTPRequest{&w, r, make(chan bool)}
	h.server.deliveries <- req

	<-req.lock
}

func NewTHTTPServer(listenAddr string) (*THTTPServer, error) {
	thriftServer := &THTTPServer{deliveries: make(chan *THTTPRequest)}

	thriftServer.server = &http.Server{
		Addr:    listenAddr,
		Handler: &HTTPHandler{thriftServer},
	}

	return thriftServer, nil
}

func (p *THTTPServer) Listen() error {
	go p.server.ListenAndServe()

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
	return nil
}

func (p *THTTPServer) Interrupt() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.interrupted = true
	return nil
}
