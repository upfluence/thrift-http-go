package http_thrift

import (
	"bytes"
	"github.com/upfluence/thrift/lib/go/thrift"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
)

type THTTPConn struct {
	conn        *httputil.ServerConn
	writeBuffer *ClosingBuffer
	req         *http.Request
}
type ClosingBuffer struct {
	*bytes.Buffer
}

func (cb *ClosingBuffer) Close() (err error) {
	return
}

func NewTHTTPConn(c *httputil.ServerConn) *THTTPConn {
	return &THTTPConn{c, &ClosingBuffer{bytes.NewBuffer([]byte{})}, nil}
}

func (p *THTTPConn) Open() error {
	return nil
}

func (p *THTTPConn) IsOpen() bool {
	return true
}

func (p *THTTPConn) Close() error {
	return nil
}

func (p *THTTPConn) Read(buf []byte) (int, error) {
	var err error
	if p.req == nil {
		p.req, err = p.conn.Read()

		if err != nil {
			return 0, thrift.NewTTransportExceptionFromError(err)
		}
	}

	return p.req.Body.Read(buf)
}

func (p *THTTPConn) Write(buf []byte) (int, error) {
	return p.writeBuffer.Write(buf)
}

func (p *THTTPConn) Flush() error {
	resp := &http.Response{
		Request:       p.req,
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          p.writeBuffer,
		Header:        make(http.Header),
		ContentLength: int64(p.writeBuffer.Len()),
	}
	resp.Header.Set("Content-Length", strconv.Itoa(p.writeBuffer.Len()))
	log.Println(strconv.Itoa(p.writeBuffer.Len()))
	p.conn.Write(p.req, resp)
	return p.conn.Close()
}
