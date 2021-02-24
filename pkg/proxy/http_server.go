package proxy

import (
	"log"
	"net"
	"sync"
)

type httpServer struct {
	addr     string
	listener *net.TCPListener

	connID  uint64
	conns   map[uint64]*httpConn
	rwmutex sync.RWMutex

	closeCh chan struct{}
	closed  bool
}

type HttpServerOpt struct {
	Addr      string
	ProxyType int
}

func NewHttpServer(opt *HttpServerOpt) *httpServer {
	s := &httpServer{
		addr:     opt.Addr,
		listener: nil,

		connID: 0,
		conns:  make(map[uint64]*httpConn),

		closeCh: make(chan struct{}),
		closed:  false,
	}
	return s
}

func (httpServer *httpServer) Start() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", httpServer.addr)
	if err != nil {
		return err
	}
	httpServer.listener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}
	go httpServer.serveAccept()
	return nil
}

func (httpServer *httpServer) serveAccept() {
	for {
		select {
		case <-httpServer.closeCh:
			return
		default:
		}

		conn, err := httpServer.listener.AcceptTCP()
		if err != nil {
			log.Printf("[E] accept conn failed. err:%s", err)
			continue
		}

		curConnId := httpServer.connID
		tcpConn := newConn(&httpConnOpt{
			Server: httpServer,
			ID:     curConnId,
			conn:   conn,
		})
		err = tcpConn.start()
		if err != nil {
			log.Printf("[E] tcp conn start faild. err:%s", err)
			continue
		}

		httpServer.rwmutex.Lock()
		httpServer.conns[curConnId] = tcpConn
		httpServer.connID++
		httpServer.rwmutex.Unlock()
	}
}

func (httpServer *httpServer) Close() error {
	httpServer.rwmutex.Lock()
	defer httpServer.rwmutex.Unlock()

	if httpServer.closed {
		return nil
	}

	httpServer.closed = true
	close(httpServer.closeCh)
	for cID := range httpServer.conns {
		delete(httpServer.conns, cID)
	}
	return httpServer.listener.Close()
}

func (httpServer *httpServer) delConn(connID uint64) {
	httpServer.rwmutex.Lock()
	defer httpServer.rwmutex.Unlock()
	if _, ok := httpServer.conns[connID]; !ok {
		return
	}
	delete(httpServer.conns, connID)
}
