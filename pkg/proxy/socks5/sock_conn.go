package socks5

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SockAuthImpl interface {
	Auth(user, pass string) bool
}

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
		log.Printf("[E] hand shake failed. err:%v", err)
		return err
	}
	log.Printf("[I] hand shake success. authType:%v", sockConn.authType)

	//服务端如需认证,则走认证流程
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
		log.Printf("[E] handleProxyInstruction failed. err:%v", err)
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
	resp.Ver = DEFAULT_VERSION
	for _, method := range req.Methods {
		if method == NMETHODS_NONE && sockConn.auth == nil {
			resp.Method = NMETHODS_NONE
			break
		}
		if method == NMETHODS_USERPASS && sockConn.auth != nil {
			resp.Method = NMETHODS_USERPASS
			break
		}
	}

	_, err = resp.WriteHandShake(sockConn.srcConn)
	if err != nil {
		log.Printf("[E] write failed. addr:%v, connID:%v, err:%v", sockConn.srcConn.RemoteAddr(), sockConn.connID, err)
		return err
	}

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
		log.Printf("[E] read auth req failed. err:%v", err)
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
		log.Printf("[E] createDstAddrConn failed. err:%v", err)
		return err
	}

	connFunc := func() (ipv4 []byte, port []byte, err error) {
		var localAddr string

		if strings.Contains(sockConn.dstNetwork, "tcp") {
			sockConn.dstConn, err = net.DialTCP(sockConn.dstNetwork, nil, sockConn.dstAddr.(*net.TCPAddr))
			if err != nil {
				log.Printf("[E] dial dst conn failed. dstNetwork:%v, dstAddr:%v, err:%v", sockConn.dstNetwork, sockConn.dstAddr.String(), err)
				return nil, nil, err
			}

			localAddr = sockConn.dstConn.LocalAddr().String()
		}

		if strings.Contains(sockConn.dstNetwork, "udp") {
			sockConn.dstConn, err = net.ListenUDP(sockConn.dstNetwork, nil)
			if err != nil {
				log.Printf("[E] listen udp failed. err:%v", err)
				return nil, nil, err
			}

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

		portUint64, _ := strconv.ParseUint(portStr, 10, 16)
		portUint16 := uint16(portUint64)
		portBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(portBytes, portUint16)

		return []byte(ipv4Str), portBytes, nil
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
		log.Printf("[E] resolve dstAddr failed. err:%v", err)
		return err
	}

	log.Printf("[I] 3. instruction read success. sockFrame:%+v", sockFrame)
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
				log.Printf("[E] read udp data from srcConn failed. err:%v", err)
				return
			}
			cancel()

			//write dst data
			sockConn.dstAddr, _ = net.ResolveUDPAddr(sockConn.dstNetwork, sockFrame.Dst.ADDR)
			_, err = sockConn.dstConn.(*net.UDPConn).WriteToUDP(sockFrame.data, sockConn.dstAddr.(*net.UDPAddr))
			if err != nil {
				log.Printf("[E] write udp data to dstConn failed. err:%v", err)
				return
			}
		} else {
			_, err := io.Copy(sockConn.dstConn, sockConn.srcConn)
			if err != nil {
				log.Printf("[E] socks5 tcp src->dst failed. err:%v", err)
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
			log.Printf("[E] socks5 tcp dst->src failed. err:%v, dstNetwork:%v, dstAddr:%v", err, sockConn.dstNetwork, sockConn.dstAddr)
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
			log.Printf("[I] close dst sockConn failed. err:%s", err)
		}
	}

	return sockConn.srcConn.Close()
}
