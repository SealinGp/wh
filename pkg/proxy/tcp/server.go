package tcp

import (
	"log"
	"net"
	"sync"
	"time"
)

type Server struct {
	addr     string
	listener *net.TCPListener

	connID  uint64
	conns   map[uint64]*Conn
	rwmutex sync.RWMutex

	proxyType int
	closeCh   chan struct{}
	closed    bool
}

type ServerOpt struct {
	Addr      string
	ProxyType int
}

func NewServer(opt *ServerOpt) *Server {
	if opt.ProxyType == UNKOWN_PROXY {
		opt.ProxyType = HTTP_PROXY
	}

	s := &Server{
		addr:     opt.Addr,
		listener: nil,

		connID: 0,
		conns:  make(map[uint64]*Conn),

		closeCh:   make(chan struct{}),
		closed:    false,
		proxyType: opt.ProxyType,
	}
	return s
}

func (server *Server) Start() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", server.addr)
	if err != nil {
		return err
	}
	server.listener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}
	go server.serveAccept()
	return nil
}

func (server *Server) serveAccept() {
	for {
		select {
		case <-server.closeCh:
			return
		default:
		}

		conn, err := server.listener.AcceptTCP()
		if err != nil {
			log.Printf("[E] accept conn failed. err:%s", err)
			continue
		}

		err = conn.SetKeepAlive(true)
		if err != nil {
			log.Printf("[E] conn set keep alive failed. err:%s", err)
			continue
		}
		err = conn.SetKeepAlivePeriod(30 * time.Minute)
		if err != nil {
			log.Printf("[E] conn set keep alive period failed. err:%s", err)
			continue
		}

		curConnId := server.connID
		tcpConn := newConn(&ConnOpt{
			Server: server,
			ID:     curConnId,
			conn:   conn,
		})
		go tcpConn.serveRead()

		server.rwmutex.Lock()
		server.conns[curConnId] = tcpConn
		server.connID++
		server.rwmutex.Unlock()
	}
}

func (server *Server) Close() error {
	server.rwmutex.Lock()
	defer server.rwmutex.Unlock()

	if server.closed {
		return nil
	}

	server.closed = true
	close(server.closeCh)
	for cID, c := range server.conns {
		err := c.Close()
		if err != nil {
			log.Printf("[E] conns close failed")
		}
		delete(server.conns, cID)
	}
	return server.listener.Close()
}

func (server *Server) delConn(connID uint64) {
	server.rwmutex.Lock()
	defer server.rwmutex.Unlock()
	if _, ok := server.conns[connID]; !ok {
		return
	}
	delete(server.conns, connID)
}
