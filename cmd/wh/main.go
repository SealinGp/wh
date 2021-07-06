package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	c_log "github.com/SealinGp/go-lib/c-log"
	"github.com/SealinGp/go-socks5"
	"github.com/SealinGp/wh/config"
	http_proxy "github.com/SealinGp/wh/proxy/http"
	"github.com/SealinGp/wh/svrs"
)

var (
	_version_ string
	c         string
	e         string
)

func main() {
	flag.StringVar(&c, "c", "config/wh.yml", "config path")
	flag.StringVar(&e, "e", "dev", "environment dev|prod")
	flag.Parse()

	//config init
	err := config.Init(c)
	if err != nil {
		c_log.E("config init failed. err:%v", err)
		return
	}

	//log init
	logLevel := c_log.LEVEL_ERR
	logPath := config.GetOptions().LogPath
	if e == "dev" {
		logPath = ""
		logLevel = c_log.LEVEL_INFO
	}
	c_log.CLogInit(&c_log.CLogOptions{
		Flag:     log.Lshortfile | log.Ltime,
		Path:     logPath,
		LogLevel: logLevel,
	})

	svrs.Init()
	defer svrs.CloseAll()

	//start proxy
	err = StartAllProxy()
	if err != nil {
		c_log.E("proxy start failed. err:%v", err)
		return
	}
	c_log.I("wh start success. version:%v, pid:%v", _version_, os.Getpid())

	//wait signal for release resources
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	sig := <-sigCh

	c_log.I("wh exit. signal:%v", sig)
}

func StartAllProxy() error {
	err := startHTTP()
	if err != nil {
		return err
	}

	err = startSocks5()
	if err != nil {
		return err
	}

	return nil
}

func startHTTP() error {
	//http server proxy
	for _, addr := range config.GetOptions().HttpProxyAddrs {
		_, _, err := net.SplitHostPort(addr)
		if err != nil {
			c_log.E("unvalid http proxy addr. addr:%v, err:%v", addr, err)
			continue
		}

		httpProxyServer := http_proxy.NewServer(&http_proxy.ServerOpt{
			Addr: addr,
		})
		if err := httpProxyServer.Start(); err != nil {
			c_log.E("http(s) start failed. addr:%v, err:%s", addr, err)
			return err
		}

		c_log.I("http proxy success. addr:%v", addr)
		_ = svrs.Add(addr, addr, httpProxyServer)
	}
	return nil
}

func startSocks5() error {
	//socks5 proxy
	for _, addr := range config.GetOptions().SocksProxyAddrs {
		_, _, err := net.SplitHostPort(addr)
		if err != nil {
			c_log.E("unvalid socks5 proxy addr ignored. addr:%v, err:%v", addr, err)
			continue
		}

		cfg := &socks5.Config{
			Logger: &LoggerWrap{},
		}
		socks5Server, err := socks5.New(cfg)
		if err != nil {
			c_log.E("new socks5 failed. err:%v", err)
			return err
		}

		socks5Wrap := &Socks5Wrap{
			addr:   addr,
			Server: socks5Server,
		}

		if err := socks5Wrap.Start(); err != nil {
			c_log.E("socks5 proxy server start failed. addr:%v, err:%s", addr, err)
			return err
		}

		c_log.I("socks5 proxy success. addr:%v", addr)
		_ = svrs.Add(addr, addr, socks5Wrap)
	}
	return nil
}

//logger wrap for socks5
type LoggerWrap struct {
}

func (l *LoggerWrap) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

//socks5wrap for socks5
type Socks5Wrap struct {
	addr string
	net.Listener
	*socks5.Server
}

func (socks5Wrap *Socks5Wrap) Start() error {
	l, err := net.Listen("tcp", socks5Wrap.addr)
	if err != nil {
		c_log.E("start socks5 proxy failed. addr:%v, err:%v", socks5Wrap.addr, err)
		return err
	}
	socks5Wrap.Listener = l

	go func() {
		socks5Wrap.Serve(l)
	}()
	return nil
}

func (socks5Wrap *Socks5Wrap) Close() error {
	return socks5Wrap.Listener.Close()
}
