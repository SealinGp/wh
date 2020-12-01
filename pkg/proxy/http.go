package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
	"time"
)

type HttpProxy struct {
	address  string
	listener net.Listener
	stopCh   chan struct{}
	closed   bool
}

type HttpProxyOpt struct {
	Address string
}

func NewHTTP(opt *HttpProxyOpt) *HttpProxy {
	httpProxy := &HttpProxy{
		address:  opt.Address,
		listener: nil,
		stopCh:   make(chan struct{}),
		closed:   false,
	}
	return httpProxy
}

func (httpProxy *HttpProxy) Start() error {
	var err error
	httpProxy.listener, err = net.Listen("tcp", httpProxy.address)
	if err != nil {
		return err
	}
	for {
		select {
		case <-httpProxy.stopCh:
			return nil
		default:
		}
		conn, err := httpProxy.listener.Accept()
		if err != nil {
			log.Printf("[E] accept failed")
		}

		go httpProxy.HandleConn(conn)
	}
}

func (httpProxy *HttpProxy) Close() error {
	if httpProxy.closed {
		return nil
	}

	close(httpProxy.stopCh)
	httpProxy.closed = true
	err := httpProxy.listener.Close()
	if err != nil {
		log.Printf("[E] close failed. addr:%s, err:%s", httpProxy.address, err)
	}
	return nil
}

func (httpProxy *HttpProxy) HandleConn(conn net.Conn) {
	defer conn.Close()
	addr, err := httpProxy.GetDstAddr(conn)
	if err != nil {
		log.Printf("[E] GetDstAddr failed. err:%s", err)
		return
	}

	dstConn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("[E] dial dst failed. err:%s, addr:%s", err, addr)
		return
	}
	defer dstConn.Close()

	//dst server -> proxy server
	go func() {
		_, err = io.Copy(conn, dstConn)
		if err != nil {
			log.Printf("[E] dial dst failed. err:%s, addr:%s", err, addr)
			return
		}
	}()

	//proxy server -> dst server
	_, err = io.Copy(dstConn, conn)
	if err != nil {
		log.Printf("[E] dial dst failed. err:%s, addr:%s", err, addr)
		return
	}
}

func (httpProxy *HttpProxy) GetDstAddr(conn net.Conn) (string, error) {
	var method, host, addr string

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	errCh := make(chan error, 1)
	input := make([]byte, 1024)
	var n int
	var err error
	go func() {
		defer close(errCh)
		n, err = conn.Read(input)
		select {
		case errCh <- err:
		case <-ctx.Done():
		}
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-errCh:
		if err != nil {
			return "", err
		}
	}

	//header
	sepIndex := bytes.IndexByte(input[:n], '\n')
	_, _ = fmt.Sscanf(string(input[:sepIndex]), "%s%s", &method, &host)
	httpUrl, err := url.Parse(host)
	if err != nil {
		return "", err
	}
	if httpUrl.Opaque == "443" {
		addr = httpUrl.Scheme + ":443"
	} else {
		addr = httpUrl.Host
		if strings.Index(httpUrl.Host, ":") == -1 {
			addr += ":80"
		}
	}

	return addr, nil
}
