package http_thrift

import (
	"net"
	"net/http/httputil"
	"sync"

	"github.com/upfluence/thrift/lib/go/thrift"
)

type THTTPServer struct {
	addr     net.Addr
	listener net.Listener

	mu          sync.RWMutex
	interrupted bool
}

func NewTHTTPServer(listenAddr string) (*THTTPServer, error) {
	addr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	return &THTTPServer{addr: addr}, nil
}

func (p *THTTPServer) Listen() error {
	if p.IsListening() {
		return nil
	}
	l, err := net.Listen(p.addr.Network(), p.addr.String())
	if err != nil {
		return err
	}
	p.listener = l
	return nil
}

func (p *THTTPServer) Accept() (thrift.TTransport, error) {
	conn, err := p.listener.Accept()

	if err != nil {
		return nil, thrift.NewTTransportExceptionFromError(err)
	}

	return NewTHTTPConn(httputil.NewServerConn(conn, nil)), nil
}

func (p *THTTPServer) Open() error {
	return p.Listen()
}

func (p *THTTPServer) Close() error {
	defer func() {
		p.listener = nil
	}()
	if p.IsListening() {
		return p.listener.Close()
	}
	return nil
}

func (p *THTTPServer) Interrupt() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.interrupted = true
	return nil
}

func (p *THTTPServer) IsListening() bool {
	return p.listener != nil
}
