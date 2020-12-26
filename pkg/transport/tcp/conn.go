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

/**
todo 1:dst->src read failed
todo 2:

20/12/09 22:14:46 [I] dst conn established. src:127.0.0.1:57118, dst:113.96.181.217:443
2020/12/09 22:14:46 [I] dst conn established. src:127.0.0.1:57119, dst:183.2.200.238:443
2020/12/09 22:15:16 [E] dst->src read from tcpConn failed. err:read tcp 192.168.1.21:57121->183.2.200.238:443: use of closed network connection
2020/12/09 22:15:16 [E] dst->src read from tcpConn failed. err:read tcp 192.168.1.21:57120->113.96.181.217:443: use of closed network connection
*/
const (
	UNKOWN_PROXY = iota
	HTTP_PROXY
)

type Conn struct {
	parentServer *Server
	ID           uint64
	dstHost      string

	writerCh chan []byte
	readerCh chan []byte

	conn     *net.TCPConn
	dst      *dstConn
	closed   bool
	closeCh  chan struct{}
	rwmutext sync.RWMutex
}

type dstConn struct {
	conn *net.TCPConn
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

		writerCh: make(chan []byte),
		readerCh: make(chan []byte),

		conn:    opt.conn,
		closed:  false,
		closeCh: make(chan struct{}),
		dst: &dstConn{
			conn: nil,
		},
	}
	return conn
}

func (tcpConn *Conn) start() error {
	err := tcpConn.createDstConn()
	if err != nil {
		return err
	}

	go tcpConn.serveReadSrc()
	go tcpConn.serveWriteDst()

	go tcpConn.serverReadDst()
	go tcpConn.serverWriteSrc()
	return nil
}

func (tcpConn *Conn) serveReadSrc() {
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
		n, err := tcpConn.conn.Read(data)
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				time.Sleep(time.Millisecond * 100)
				continue
			}
			if err == io.EOF {
				return
			}
			log.Printf("[E] serveReadSrc read failed. err:%s", err)
			continue
		}
		tcpConn.readerCh <- data[:n]

		log.Printf("[I] read connID:%d, data:%s ", tcpConn.ID, data[:n])
	}
}

func (tcpConn *Conn) serveWriteDst() {
	for {
		data, ok := <-tcpConn.readerCh
		if !ok {
			return
		}

		_, err := tcpConn.dst.conn.Write(data)
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				time.Sleep(time.Millisecond * 100)
				continue
			}
			if err == io.EOF {
				return
			}
			log.Printf("[E] serveWriteDst read failed. err:%s", err)
		}
	}

}

func (tcpConn *Conn) serverReadDst() {
	defer func() {
		err := tcpConn.Close()
		if err != nil {
			log.Printf("[E] close conn failed. connID:%d, err:%s", tcpConn.ID, err)
			return
		}
		tcpConn.parentServer.delConn(tcpConn.ID)
	}()
	for {
		data := make([]byte, 1500)
		n, err := tcpConn.dst.conn.Read(data)
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				time.Sleep(time.Millisecond * 100)
				continue
			}
			if err == io.EOF {
				return
			}
			log.Printf("[E] serverReadDst read failed. err:%s", err)
			continue
		}

		tcpConn.writerCh <- data[:n]
	}
}

func (tcpConn *Conn) serverWriteSrc() {
	for {
		data, ok := <-tcpConn.writerCh
		if !ok {
			return
		}

		_, err := tcpConn.conn.Write(data)
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				time.Sleep(time.Millisecond * 100)
				continue
			}
			if err == io.EOF {
				return
			}
			log.Printf("[E] serverWriteSrc failed. err:%s", err)
		}
	}
}

func (tcpConn *Conn) createDstConn() error {
	if tcpConn.parentServer.proxyType == HTTP_PROXY {
		req, err := http.ReadRequest(bufio.NewReader(tcpConn.conn))
		if err != nil {
			return err
		}

		if req.Method != http.MethodConnect {
			return errors.New("wrong proxy method")
		}

		if req.Header.Get("Proxy-Connection") == "keep-alive" {
			if err = tcpConn.conn.SetKeepAlive(true); err != nil {
				return err
			}
			if err = tcpConn.conn.SetKeepAlivePeriod(time.Minute * 30); err != nil {
				return err
			}
		}

		//1.create dst conn
		tcpConn.dstHost = req.Host
		dstTcpAddr, err := net.ResolveTCPAddr("tcp", req.Host)
		if err != nil {
			return err
		}
		conn, err := net.DialTCP("tcp", nil, dstTcpAddr)
		if err != nil {
			return err
		}
		if req.Header.Get("Proxy-Connection") == "keep-alive" {
			if err = tcpConn.conn.SetKeepAlive(true); err != nil {
				return err
			}
			if err = tcpConn.conn.SetKeepAlivePeriod(time.Minute * 30); err != nil {
				return err
			}
			if err = conn.SetKeepAlive(true); err != nil {
				return err
			}
			if err = conn.SetKeepAlivePeriod(time.Minute * 30); err != nil {
				return err
			}
		}
		tcpConn.dst.conn = conn

		//2.send http ok to tcpConn.Writer
		inReq := &http.Response{
			Status:     "200 Connection Established",
			StatusCode: http.StatusOK,
			Proto:      req.Proto,
			ProtoMajor: req.ProtoMajor,
			ProtoMinor: req.ProtoMinor,
			Request:    req,
			TLS:        req.TLS,
		}
		return inReq.Write(tcpConn.conn)
	}

	return errors.New("proxy type not supported yet")
}

func (tcpConn *Conn) Close() error {
	tcpConn.rwmutext.Lock()
	defer tcpConn.rwmutext.Unlock()

	if tcpConn.closed {
		return nil
	}

	close(tcpConn.closeCh)
	close(tcpConn.writerCh)
	close(tcpConn.readerCh)
	tcpConn.closed = true
	if tcpConn.dst != nil && tcpConn.dst.conn != nil {
		err := tcpConn.dst.conn.Close()
		if err != nil {
			log.Printf("[I] close dst conn failed. err:%s", err)
		}
	}
	return tcpConn.conn.Close()
}
