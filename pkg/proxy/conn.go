package tcp

import (
	"bufio"
	"io"
	"log"
	"net"
	"sync"
)

type Conn struct {
	parentServer *Server
	ID           uint64
	writer       *bufio.Writer
	reader       *bufio.Reader
	conn         *net.TCPConn
	dataPool     *sync.Pool
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
		dataPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 1500)
			},
		},
	}
	return conn
}

func (tcpConn *Conn) serveRead() {
	defer func() {
		err := tcpConn.conn.Close()
		if err != nil {
			log.Printf("[E] close conn failed. connID:%d, err:%s", tcpConn.ID, err)
		}
		tcpConn.parentServer.DelConn(tcpConn.ID)
	}()

	for {
		data := tcpConn.dataPool.Get().([]byte)
		n, err := tcpConn.reader.Read(data)
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Printf("[E] read from tcpConn failed. err:%s", err)
			continue
		}

		//CONNECT
		//get dst addr
		dstAddr := string(data[:n])
		dstaddr, err := net.ResolveTCPAddr("tcp", dstAddr)
		if err != nil {
			log.Printf("[E] resolve dst addr failed. err:%s", err)
			return
		}
		dstLis, err := net.DialTCP("tcp", nil, dstaddr)

		tcpConn.dataPool.Put(data[:0])
	}
}
