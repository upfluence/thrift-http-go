package http_thrift

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/upfluence/thrift/lib/go/thrift"
)

var errTransportClosed = errors.New("thrift/transport/http: Transport closed")

type THTTPClient struct {
	response      *http.Response
	url           *url.URL
	requestBuffer *bytes.Buffer
	retries       uint
	retryDelay    time.Duration
	timeout       time.Duration
}

type THTTPClientTransportFactory struct {
	url        string
	retries    uint
	retryDelay time.Duration
	timeout    time.Duration
}

func (p *THTTPClientTransportFactory) GetTransport(trans thrift.TTransport) thrift.TTransport {
	if trans != nil {
		t, ok := trans.(*THTTPClient)
		if ok && t.url != nil {
			t2, _ := NewTHTTPClient(t.url.String(), t.retries, t.retryDelay, t.timeout)
			return t2
		}
	}

	s, _ := NewTHTTPClient(p.url, p.retries, p.retryDelay, p.timeout)

	return s
}

func NewTHTTPClientTransportFactory(url string, retries uint, retryDelay, timeout time.Duration) *THTTPClientTransportFactory {
	return &THTTPClientTransportFactory{url, retries, retryDelay, timeout}
}

func NewTHTTPClient(urlstr string, retries uint, retryDelay, timeout time.Duration) (thrift.TTransport, error) {
	parsedURL, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 0, 1024)
	return &THTTPClient{
		url:           parsedURL,
		requestBuffer: bytes.NewBuffer(buf),
		retries:       retries,
		retryDelay:    retryDelay,
		timeout:       timeout,
	}, nil
}

func (p *THTTPClient) Open() error {
	return nil
}

func (p *THTTPClient) IsOpen() bool {
	return p.requestBuffer != nil
}

func (p *THTTPClient) Close() error {
	if p.requestBuffer != nil {
		p.requestBuffer.Reset()
		p.requestBuffer = nil
	}

	return nil
}

func (p *THTTPClient) Read(buf []byte) (int, error) {
	if p.response == nil {
		return 0, thrift.NewTTransportException(
			thrift.NOT_OPEN,
			"Response buffer is empty, no request.",
		)
	}

	n, err := p.response.Body.Read(buf)
	if n > 0 && (err == nil || err == io.EOF) {
		return n, nil
	}

	return n, thrift.NewTTransportExceptionFromError(err)
}

func (p *THTTPClient) Write(buf []byte) (int, error) {
	if p.requestBuffer == nil {
		return 0, errTransportClosed
	}

	n, err := p.requestBuffer.Write(buf)
	return n, err
}

func (p *THTTPClient) doRequest() (*http.Response, error) {
	client := &http.Client{Timeout: p.timeout}
	req, err := http.NewRequest("POST", p.url.String(), p.requestBuffer)
	if err != nil {
		return nil, thrift.NewTTransportExceptionFromError(err)
	}

	req.Header.Add("Content-Type", "application/x-thrift")

	response, err := client.Do(req)

	if err != nil {
		return nil, thrift.NewTTransportExceptionFromError(err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, thrift.NewTTransportException(
			thrift.UNKNOWN_TRANSPORT_EXCEPTION,
			"HTTP Response code: "+strconv.Itoa(response.StatusCode),
		)
	}

	return response, nil
}

func (p *THTTPClient) Flush() error {
	var (
		err error
		i   uint = 1
	)

	if p.requestBuffer == nil {
		return errTransportClosed
	}

	p.response, err = p.doRequest()

	for err != nil && i < p.retries {
		p.response, err = p.doRequest()

		if err != nil {
			i++
			time.Sleep(p.retryDelay)
			log.Println(err.Error())
		}
	}

	p.requestBuffer.Reset()
	return err
}
