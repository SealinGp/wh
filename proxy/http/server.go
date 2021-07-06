package http_proxy

import (
	"net"
	"sync"

	c_log "github.com/SealinGp/go-lib/c-log"
)

type Server struct {
	addr     string
	listener *net.TCPListener

	connID  uint64
	conns   map[uint64]*conn
	rwmutex sync.RWMutex

	closeCh chan struct{}
	closed  bool
}

type ServerOpt struct {
	Addr      string
	ProxyType int
}

func NewServer(opt *ServerOpt) *Server {
	httpServer := &Server{
		addr:     opt.Addr,
		listener: nil,

		connID: 0,
		conns:  make(map[uint64]*conn),

		closeCh: make(chan struct{}),
		closed:  false,
	}
	return httpServer
}

func (httpServer *Server) Start() error {
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

func (httpServer *Server) serveAccept() {
	for {
		select {
		case <-httpServer.closeCh:
			return
		default:
		}

		conn, err := httpServer.listener.AcceptTCP()
		if err != nil {
			c_log.E("accept conn failed. err:%s", err)
			continue
		}

		curConnId := httpServer.connID
		tcpConn := newConn(&connOpt{
			Server: httpServer,
			ID:     curConnId,
			conn:   conn,
		})
		err = tcpConn.start()
		if err != nil {
			//不打印非隧道的一次性代理请求
			if err == ErrNotTunnelProxy {
				continue
			}
			c_log.E("tcp conn start failed. err:%s", err)
			continue
		}

		httpServer.rwmutex.Lock()
		httpServer.conns[curConnId] = tcpConn
		httpServer.connID++
		httpServer.rwmutex.Unlock()
	}
}

func (httpServer *Server) Close() error {
	httpServer.rwmutex.Lock()
	defer httpServer.rwmutex.Unlock()

	if httpServer.closed {
		return nil
	}

	httpServer.closed = true
	close(httpServer.closeCh)
	for connID, conn := range httpServer.conns {
		err := conn.Close()
		if err != nil {
			c_log.E("close conn failed. connID:%v, err:%v", connID, err)
		}
		delete(httpServer.conns, connID)
	}
	return httpServer.listener.Close()
}

func (httpServer *Server) delConn(connID uint64) {
	httpServer.rwmutex.Lock()
	defer httpServer.rwmutex.Unlock()
	if _, ok := httpServer.conns[connID]; !ok {
		return
	}
	delete(httpServer.conns, connID)
}
