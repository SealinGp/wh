package tcp

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
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
	dst          *dstConn
	closed       bool
	closeCh      chan struct{}
	rwmutext     sync.RWMutex
}

type dstConn struct {
	conn   *net.TCPConn
	writer *bufio.Writer
	reader *bufio.Reader
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
		dst: &dstConn{
			writer: nil,
			reader: nil,
		},
	}
	return conn
}

func (tcpConn *Conn) serveRead() {
	err := tcpConn.dstConnInit()
	if err != nil {
		log.Printf("[E] http req failed. err:%s", err)
		return
	}
	log.Printf("[I] dst conn established. src:%s, dst:%s", tcpConn.conn.RemoteAddr(), tcpConn.dst.conn.RemoteAddr())

	go tcpConn.dstToSrc()
	go tcpConn.srcToDst()
}

func (tcpConn *Conn) dstConnInit() error {
	tcpConn.rwmutext.Lock()
	defer tcpConn.rwmutext.Unlock()

	if tcpConn.parentServer.proxyType == HTTP_PROXY {
		req, err := http.ReadRequest(tcpConn.reader)
		if err != nil {
			return err
		}

		if req.Method != http.MethodConnect {
			return errors.New("wrong proxy method")
		}

		//1.create dst conn
		dstTcpAddr, err := net.ResolveTCPAddr("tcp", req.Host)
		if err != nil {
			return err
		}
		conn, err := net.DialTCP("tcp", nil, dstTcpAddr)
		if err != nil {
			return err
		}
		tcpConn.dst.conn = conn
		tcpConn.dst.reader = bufio.NewReader(conn)
		tcpConn.dst.writer = bufio.NewWriter(conn)

		//2.send http ok to tcpConn.Writer
		inReq := &http.Response{
			Status:     "200 Connection Established",
			StatusCode: http.StatusOK,
			Proto:      req.Proto,
			ProtoMajor: req.ProtoMajor,
			ProtoMinor: req.ProtoMinor,
		}
		return inReq.Write(tcpConn.writer)
	}

	return errors.New("proxy type not supported yet")
}

func (tcpConn *Conn) srcToDst() {
	defer func() {
		err := tcpConn.Close()
		if err != nil {
			log.Printf("[E] close conn failed. connID:%d, err:%s", tcpConn.ID, err)
			return
		}
		tcpConn.parentServer.delConn(tcpConn.ID)
	}()

	for {
		select {
		case <-tcpConn.closeCh:
			return
		default:
		}

		data := make([]byte, 1500)
		n, err := tcpConn.reader.Read(data)
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				time.Sleep(time.Millisecond)
				continue
			}
			if err == io.EOF {
				return
			}
			log.Printf("[E] read from tcpConn failed. err:%s", err)
			continue
		}

		_, err = tcpConn.dst.writer.Write(data[:n])
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				time.Sleep(time.Millisecond)
				continue
			}
			if err == io.EOF {
				return
			}
			log.Printf("[E] write to dst tcpConn failed. err:%s", err)
			continue
		}
	}
}

func (tcpConn *Conn) dstToSrc() {
	defer func() {
		err := tcpConn.Close()
		if err != nil {
			log.Printf("[E] close conn failed. connID:%d, err:%s", tcpConn.ID, err)
			return
		}
		tcpConn.parentServer.delConn(tcpConn.ID)
	}()

	for {
		select {
		case <-tcpConn.closeCh:
			return
		default:
		}

		data := make([]byte, 1500)
		n, err := tcpConn.dst.reader.Read(data)
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				time.Sleep(time.Millisecond)
				continue
			}
			if err == io.EOF {
				return
			}
			log.Printf("[E] read from tcpConn failed. err:%s", err)
			continue
		}

		_, err = tcpConn.writer.Write(data[:n])
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				time.Sleep(time.Millisecond)
				continue
			}
			if err == io.EOF {
				return
			}
			log.Printf("[E] write to dst tcpConn failed. err:%s", err)
			continue
		}
	}
}

func (tcpConn *Conn) Close() error {
	tcpConn.rwmutext.Lock()
	defer tcpConn.rwmutext.Unlock()

	if tcpConn.closed {
		return nil
	}

	close(tcpConn.closeCh)
	tcpConn.closed = true
	err := tcpConn.dst.conn.Close()
	if err != nil {
		log.Printf("[I] close dst conn failed. err:%s", err)
	}
	return tcpConn.conn.Close()
}
