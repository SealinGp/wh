package tcp

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
)

const (
	UNKOWN_PROXY = iota
	HTTP_PROXY
)

type Conn struct {
	parentServer *Server
	ID           uint64
	writer       *bufio.Writer
	reader       *bufio.Reader
	conn         *net.TCPConn
	dstServer    map[string]*Server
	closed       bool
	closeCh      chan struct{}
}

type ConnOpt struct {
	Server *Server
	ID     uint64
	conn   *net.TCPConn
}

func newConn(opt *ConnOpt) *Conn {
	conn := &Conn{
		parentServer: opt.Server,
		ID:           opt.ID,
		writer:       bufio.NewWriter(opt.conn),
		reader:       bufio.NewReader(opt.conn),
		conn:         opt.conn,
		closed:       false,
		closeCh:      make(chan struct{}),
	}
	return conn
}

func (tcpConn *Conn) serveRead() {
	defer func() {
		err := tcpConn.conn.Close()
		if err != nil {
			log.Printf("[E] close conn failed. connID:%d, err:%s", tcpConn.ID, err)
		}
		tcpConn.parentServer.delConn(tcpConn.ID)
	}()

	for {
		if tcpConn.parentServer.proxyType == HTTP_PROXY {
			err := tcpConn.httpRead()
			if err != nil {
				if err == io.EOF {
					log.Printf("[I] received io.EOF. err:%s", err)
					return
				}
				log.Printf("[E] http req failed. err:%s", err)
			}
			continue
		}

		data := make([]byte, 1500)
		n, err := tcpConn.reader.Read(data)
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Printf("[E] read from tcpConn failed. err:%s", err)
			continue
		}
		log.Printf("[I] tcp conn recevied. connID:%d data:%s", tcpConn.ID, string(data[:n]))
	}
}

func (tcpConn *Conn) httpRead() error {
	req, err := http.ReadRequest(tcpConn.reader)
	if err != nil {
		return err
	}

	if req.Method != http.MethodConnect {
		return errors.New("wrong proxy method")
	}

	log.Printf("[I] http req received. host:%s", req.Host)
	return nil
}
