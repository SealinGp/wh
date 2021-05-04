package http_svr

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	c_log "github.com/SealinGp/go-lib/c-log"
)

type httpConn struct {
	parentServer *HttpServer
	ID           uint64
	srcConn      *net.TCPConn
	dstConn      *net.TCPConn

	closed  bool
	closeCh chan struct{}
	rwmutex sync.RWMutex
}

type httpConnOpt struct {
	Server *HttpServer
	ID     uint64
	conn   *net.TCPConn
}

func newConn(opt *httpConnOpt) *httpConn {
	httpConn := &httpConn{
		parentServer: opt.Server,
		ID:           opt.ID,
		srcConn:      opt.conn,
		dstConn:      nil,

		closed:  false,
		closeCh: make(chan struct{}),
	}
	return httpConn
}

func (tcpConn *httpConn) start() error {
	err := tcpConn.createDstConn()
	if err != nil {
		_ = tcpConn.Close()
		return err
	}

	go tcpConn.dstToSrc()
	go tcpConn.srcToDst()
	return nil
}

func (tcpConn *httpConn) dstToSrc() {
	defer func() {
		err := tcpConn.Close()
		if err != nil {
			c_log.E("close httpConn failed. connID:%d, err:%v", tcpConn.ID, err)
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

		_, err := io.Copy(tcpConn.srcConn, tcpConn.dstConn)
		if err != nil {
			c_log.E("dst->src failed. err:%v", err)
			return
		}
	}
}

func (tcpConn *httpConn) srcToDst() {
	defer func() {
		err := tcpConn.Close()
		if err != nil {
			c_log.E("close httpConn failed. connID:%d, err:%v", tcpConn.ID, err)
			return
		}
		tcpConn.parentServer.delConn(tcpConn.ID)
	}()

	for {
		_, err := io.Copy(tcpConn.dstConn, tcpConn.srcConn)
		if err != nil {
			c_log.E("src->dst failed. err:%s", err)
			return
		}
	}
}

func (tcpConn *httpConn) createDstConn() error {
	srcReq, err := http.ReadRequest(bufio.NewReader(tcpConn.srcConn))
	if err != nil {
		return err
	}

	srcReq.Response = &http.Response{
		Header: make(http.Header),
	}

	//非隧道代理,则直接转发请求
	if srcReq.Method != http.MethodConnect {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		//构建dst http request
		dstReq, err := http.NewRequestWithContext(ctx, srcReq.Method, srcReq.URL.String(), srcReq.Body)
		if err != nil {
			c_log.E("new dst req failed. err:%v", err)
			return err
		}

		srcIP, _, err := net.SplitHostPort(tcpConn.srcConn.RemoteAddr().String())
		if err != nil {
			c_log.E("get src addr failed. err:%v", err)
			return err
		}

		//添加转发ip头,X-Forward-For,意思是为哪个客户端转发的请求,逗号隔开
		dstReq.Header = srcReq.Header.Clone()
		forwardForHeader := dstReq.Header.Get("X-Forward-For")
		if forwardForHeader != "" {
			forwardForHeaders := strings.Split(forwardForHeader, ",")
			forwardForHeaders = append(forwardForHeaders, srcIP)
			forwardForHeader = strings.Join(forwardForHeaders, ",")
		} else {
			forwardForHeader = srcIP
		}
		dstReq.Header.Set("X-Forward-For", forwardForHeader)

		//代理请求dst端
		dstResp, err := http.DefaultClient.Do(dstReq)
		if err != nil {
			c_log.E("do dst req failed. err:%v", err)
			return err
		}

		//返回结果给src端
		for k, vs := range dstResp.Header {
			for _, v1 := range vs {
				srcReq.Response.Header.Add(k, v1)
			}
		}
		srcReq.Response.StatusCode = dstResp.StatusCode
		srcReq.Response.Body = dstResp.Body

		log.Printf("[D] srcIP:%v dstIP:%v X-Forward-For:%v", tcpConn.srcConn.RemoteAddr().String(), srcReq.Host, forwardForHeader)
		err = srcReq.Response.Write(tcpConn.srcConn)
		if err != nil {
			c_log.E("dst resp write to src resp failed. err:%v", err)
			return err
		}
		return ErrNotTunnelProxy
	}

	//隧道代理,在建立dst连接后需要返回Connection Established 给src client,然后持续互相转发
	dstIP, dstPort, _ := net.SplitHostPort(srcReq.Host)
	if dstPort == "" {
		dstPort = "80"
	}

	dstAddr := net.JoinHostPort(dstIP, dstPort)
	dstTcpAddr, err := net.ResolveTCPAddr("tcp", dstAddr)
	if err != nil {
		return err
	}

	dstConn, err := net.DialTCP("tcp", nil, dstTcpAddr)
	if err != nil {
		return err
	}

	tcpConn.dstConn = dstConn
	keepAlive := srcReq.Header.Get("Proxy-Connection") == "keep-alive"
	if keepAlive {
		err = func() error {
			var keepAliveErr error
			keepAliveErr = tcpConn.srcConn.SetKeepAlive(true)
			keepAliveErr = tcpConn.srcConn.SetKeepAlivePeriod(time.Minute * 30)
			keepAliveErr = tcpConn.dstConn.SetKeepAlive(true)
			keepAliveErr = tcpConn.dstConn.SetKeepAlivePeriod(time.Minute * 30)
			return keepAliveErr
		}()
		if err != nil {
			c_log.E("keep alive failed. err:%v", err)
			return err
		}
	}

	//隧道代理
	srcReq.Response.Status = "200 Connection Established"
	srcReq.Response.StatusCode = http.StatusOK
	srcReq.Response.Proto = srcReq.Proto
	srcReq.Response.ProtoMajor = srcReq.ProtoMajor
	srcReq.Response.ProtoMinor = srcReq.ProtoMinor
	srcReq.Response.Header.Set("Content-Type", "text/plain; charset=utf-8")

	log.Printf("[D] src:%v proxy:%v dst:%v", tcpConn.srcConn.RemoteAddr(), tcpConn.dstConn.LocalAddr(), tcpConn.dstConn.RemoteAddr())
	return srcReq.Response.Write(tcpConn.srcConn)
}

func (tcpConn *httpConn) Close() error {
	tcpConn.rwmutex.Lock()
	defer tcpConn.rwmutex.Unlock()

	if tcpConn.closed {
		return nil
	}

	tcpConn.closed = true
	close(tcpConn.closeCh)

	if tcpConn.dstConn != nil {
		err := tcpConn.dstConn.Close()
		if err != nil {
			c_log.I("close dst httpConn failed. err:%s", err)
		}
	}

	return tcpConn.srcConn.Close()
}
