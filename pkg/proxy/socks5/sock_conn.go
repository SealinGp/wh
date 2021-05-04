package socks5

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type sockConn struct {
	parentServer *SockServer
	connID       uint64

	srcConn    *net.TCPConn
	dstNetwork string
	dstAddr    net.Addr
	dstConn    net.Conn

	authType uint8
	auth     SockAuthImpl

	closeCh chan struct{}
	closed  bool
	rwmutex sync.RWMutex
}

type sockConnOpt struct {
	parentServer *SockServer
	conn         *net.TCPConn
	connID       uint64
	auth         SockAuthImpl
}

func newSockConn(opt *sockConnOpt) *sockConn {
	sockConn := &sockConn{
		parentServer: opt.parentServer,
		srcConn:      opt.conn,
		connID:       opt.connID,
		auth:         opt.auth,

		closeCh: make(chan struct{}),
		closed:  false,
	}

	return sockConn
}

func (sockConn *sockConn) Start() error {
	//客户端-服务端握手
	err := sockConn.handShake()
	if err != nil {
		sockConn.Close()
		log.Printf("[E] handShake failed. connID:%v, err:%v", sockConn.connID, err)
		return err
	}

	//服务端如需user_pass认证,则走认证流程
	if sockConn.authType == NMETHODS_USERPASS && sockConn.auth != nil {
		if err = sockConn.handleAuth(); err != nil {
			sockConn.Close()
			return err
		}
	}

	//开始处理客户端代理请求
	err = sockConn.handleProxyInstruction()
	if err != nil {
		sockConn.Close()
		log.Printf("[E] handleProxyInstruction failed. connID:%v, err:%v", sockConn.connID, err)
		return err
	}

	go sockConn.srcToDst()
	go sockConn.dstToSrc()

	return nil
}

//开始握手
func (sockConn *sockConn) handShake() error {
	req := NewSockFrame()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := req.ReadHandShake(sockConn.srcConn, ctx)
	if err != nil {
		log.Printf("[E] handshake read failed. addr:%v, connID:%v, err:%v", sockConn.srcConn.RemoteAddr(), sockConn.connID, err)
		return err
	}

	//服务端选择认证方式
	resp := NewSockFrame()
	resp.Method = NMETHODS_NONE
	resp.Ver = DEFAULT_VERSION
	for _, method := range req.Methods {
		if method == NMETHODS_USERPASS && sockConn.auth != nil {
			resp.Method = NMETHODS_USERPASS
			break
		}
		if method == NMETHODS_GSSAPI {
			//todo
		}
	}

	_, err = resp.WriteHandShake(sockConn.srcConn)
	if err != nil {
		log.Printf("[E] write failed. connID:%v, addr:%v, err:%v", sockConn.connID, sockConn.srcConn.RemoteAddr(), err)
		return err
	}

	log.Printf("[I] handShake finished. connID:%v", sockConn.connID)

	sockConn.authType = resp.Method
	return nil
}

//开始认证流程
func (sockConn *sockConn) handleAuth() error {
	sockFrame := NewSockFrame()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	sockAuthRequest, err := sockFrame.ReadAuthReq(ctx, sockConn.srcConn)
	if err != nil {
		log.Printf("[E] read auth req failed. connID:%v, err:%v", sockConn.connID, err)
		return err
	}

	isAuthSuccess := sockConn.auth.Auth(
		sockAuthRequest.GetUser(),
		sockAuthRequest.GetPassWord(),
	)
	_, err = sockFrame.WriteAuthResp(sockConn.srcConn, isAuthSuccess)
	return err
}

//开始处理客户端代理指令
func (sockConn *sockConn) handleProxyInstruction() error {
	err := sockConn.parseProxyInstruction()
	if err != nil {
		log.Printf("[E] createDstAddrConn failed. connID:%v, err:%v", sockConn.connID, err)
		return err
	}

	connFunc := func() (ipv4 []byte, port []byte, err error) {
		var localAddr string

		if strings.Contains(sockConn.dstNetwork, "tcp") {
			dscConn, err := net.DialTCP(sockConn.dstNetwork, nil, sockConn.dstAddr.(*net.TCPAddr))
			if err != nil {
				log.Printf("[E] dial dst conn failed. connID:%v, dstNetwork:%v, dstAddr:%v, err:%v", sockConn.connID, sockConn.dstNetwork, sockConn.dstAddr.String(), err)
				return nil, nil, err
			}

			sockConn.dstConn = dscConn
			localAddr = sockConn.dstConn.LocalAddr().String()
		}

		if strings.Contains(sockConn.dstNetwork, "udp") {
			dstConn, err := net.ListenUDP(sockConn.dstNetwork, nil)
			if err != nil {
				log.Printf("[E] listen udp failed. connID:%v, err:%v", sockConn.connID, err)
				return nil, nil, err
			}

			sockConn.dstConn = dstConn
			localAddr = sockConn.dstConn.LocalAddr().String()
		}

		if sockConn.dstConn == nil {
			return nil, nil, errors.New("unsupported dstNetwork")
		}

		//ipv4 + port
		ipv4Str, portStr, err := net.SplitHostPort(localAddr)
		if err != nil {
			return nil, nil, errors.New(fmt.Sprintf("get local host and port failed. err:%v", err))
		}

		return []byte(ipv4Str), []byte(portStr), nil
	}

	sockFrame := NewSockFrame()
	_, err = sockFrame.WriteInstruction(sockConn.srcConn, connFunc)
	return err
}

//解析客户端需要代理的信息
func (sockConn *sockConn) parseProxyInstruction() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	sockFrame := NewSockFrame()
	err := sockFrame.ReadInstruction(ctx, sockConn.srcConn)
	if err != nil {
		return err
	}

	switch sockFrame.Cmd {
	case CMD_TCP:
		sockConn.dstNetwork = CMD_MAP[CMD_TCP]
		sockConn.dstAddr, err = net.ResolveTCPAddr(sockConn.dstNetwork, sockFrame.Dst.String())
	case CMD_UDP:
		sockConn.dstNetwork = CMD_MAP[CMD_UDP]
		sockConn.dstAddr, err = net.ResolveUDPAddr(sockConn.dstNetwork, sockFrame.Dst.String())
	default:
		return errors.New(fmt.Sprintf("unsupported cmd. dstNetwork:%v", CMD_MAP[sockFrame.Cmd]))
	}

	if err != nil {
		log.Printf("[E] resolve dstAddr failed. connID:%v, err:%v", sockConn.connID, err)
		return err
	}

	log.Printf("[I] parseProxyInstruction success. connID:%v, sockFrame:%+v dstAddr:%v", sockConn.connID, sockFrame, sockConn.dstAddr)
	return nil
}

func (sockConn *sockConn) srcToDst() {
	defer func() {
		err := sockConn.Close()
		if err != nil {
			log.Printf("[E] close socks5Conn failed. connID:%d, err:%v", sockConn.connID, err)
			return
		}
		sockConn.parentServer.delConn(sockConn.connID)
	}()

	for {
		select {
		case <-sockConn.closeCh:
			return
		default:
		}

		if _, ok := sockConn.dstConn.(*net.UDPConn); ok {
			//read src data
			sockFrame := NewSockFrame()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			err := sockFrame.ReadUdpData(sockConn.srcConn, ctx)
			if err != nil {
				cancel()
				log.Printf("[E] read udp data from srcConn failed. connID:%v, err:%v", sockConn.connID, err)
				return
			}
			cancel()

			//write dst data
			sockConn.dstAddr, _ = net.ResolveUDPAddr(sockConn.dstNetwork, sockFrame.Dst.ADDR)
			_, err = sockConn.dstConn.(*net.UDPConn).WriteToUDP(sockFrame.data, sockConn.dstAddr.(*net.UDPAddr))
			if err != nil {
				log.Printf("[E] write udp data to dstConn failed. connID:%v, err:%v", sockConn.connID, err)
				return
			}
		} else {
			_, err := io.Copy(sockConn.dstConn, sockConn.srcConn)
			if err != nil {
				log.Printf("[E] socks5 tcp src->dst failed. connID:%v, err:%v", sockConn.connID, err)
				return
			}

		}
	}
}

func (sockConn *sockConn) dstToSrc() {
	defer func() {
		err := sockConn.Close()
		if err != nil {
			log.Printf("[E] close socks5Conn failed. connID:%d, err:%v", sockConn.connID, err)
			return
		}
		sockConn.parentServer.delConn(sockConn.connID)
	}()

	for {
		select {
		case <-sockConn.closeCh:
			return
		default:
		}

		_, err := io.Copy(sockConn.srcConn, sockConn.dstConn)
		if err != nil {
			log.Printf("[E] socks5 tcp dst->src failed. connID:%v, dstNetwork:%v, dstAddr:%v, err:%v", sockConn.connID, sockConn.dstNetwork, sockConn.dstAddr, err)
			return
		}
	}
}

func (sockConn *sockConn) Close() error {
	sockConn.rwmutex.Lock()
	defer sockConn.rwmutex.Unlock()

	if sockConn.closed {
		return nil
	}

	close(sockConn.closeCh)
	sockConn.closed = true

	if sockConn.dstConn != nil {
		err := sockConn.dstConn.Close()
		if err != nil {
			log.Printf("[I] close dst sockConn failed. connID:%v, err:%s", sockConn.connID, err)
		}
	}

	return sockConn.srcConn.Close()
}
