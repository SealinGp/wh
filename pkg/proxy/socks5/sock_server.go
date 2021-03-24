package socks5

import (
	"github.com/SealinGp/wh/pkg/proxy"
	"log"
	"net"
	"sync"
)

type SockServer struct {
	addr     string
	listener *net.TCPListener
	auth     SockAuthImpl

	connID  uint64
	conns   map[uint64]*sockConn
	rwmutex sync.RWMutex

	closeCh chan struct{}
	closed  bool
}

type SockServerOpt struct {
	Addr string
	Auth SockAuthImpl
}

func NewSockServer(opt *SockServerOpt) *SockServer {
	sockServer := &SockServer{
		addr: opt.Addr,
		auth: opt.Auth,

		listener: nil,
		conns:    make(map[uint64]*sockConn),

		closeCh: make(chan struct{}),
		closed:  false,
	}

	return sockServer
}

func (sockServer *SockServer) GetAddr() string {
	return sockServer.addr
}

func (sockServer *SockServer) GetType() string {
	return proxy.SOCK5
}

func (sockServer *SockServer) Start() error {
	lTcpAddr, err := net.ResolveTCPAddr("tcp", sockServer.addr)
	if err != nil {
		return err
	}

	sockServer.listener, err = net.ListenTCP("tcp", lTcpAddr)
	if err != nil {
		return err
	}

	go sockServer.serveAccept()
	return nil
}

func (sockServer *SockServer) serveAccept() {
	for {
		select {
		case <-sockServer.closeCh:
			return
		default:
		}

		conn, err := sockServer.listener.AcceptTCP()
		if err != nil {
			log.Printf("[E] accpet conn failed. err:%v", err)
			continue
		}

		sockServer.rwmutex.RLock()
		curConnId := sockServer.connID
		sockServer.rwmutex.RUnlock()

		sockConn := newSockConn(&sockConnOpt{
			sockServer,
			conn,
			curConnId,
			sockServer.auth,
		})
		err = sockConn.Start()
		if err != nil {
			log.Printf("[E] sock conn start failed. err:%v", err)
			continue
		}

		sockServer.rwmutex.Lock()
		sockServer.conns[curConnId] = sockConn
		sockServer.connID++
		sockServer.rwmutex.Unlock()
	}
}

func (sockServer *SockServer) delConn(connId uint64) {
	sockServer.rwmutex.Lock()
	defer sockServer.rwmutex.Unlock()

	delete(sockServer.conns, connId)
}

func (sockServer *SockServer) Close() error {
	sockServer.rwmutex.Lock()
	defer sockServer.rwmutex.Unlock()

	if sockServer.closed {
		return nil
	}

	for connID, conn := range sockServer.conns {
		err := conn.Close()
		if err != nil {
			log.Printf("[E] close conn failed. connID:%v, err:%v", connID, err)
		}
		delete(sockServer.conns, connID)
	}

	err := sockServer.listener.Close()
	if err != nil {
		log.Printf("[E] listener close failed. err:%v", err)
		return err
	}

	sockServer.closed = true
	close(sockServer.closeCh)
	return nil
}
