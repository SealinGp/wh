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
	connID   uint64
	conns    map[uint64]*net.TCPConn
	rwmutex  sync.RWMutex
	closeCh  chan struct{}
	closed   bool
}

type ServerOpt struct {
	Addr string
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
		server.conns[curConnId] = conn
		server.connID++
		server.rwmutex.Unlock()
	}
}

func (server *Server) Close() error {
	if server.closed {
		return nil
	}

	server.closed = true
	close(server.closeCh)
	return server.listener.Close()
}

func (server *Server) DelConn(connID uint64) {
	server.rwmutex.Lock()
	defer server.rwmutex.Unlock()
	delete(server.conns, connID)
}
