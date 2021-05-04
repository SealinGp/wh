package main

import (
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	c_log "github.com/SealinGp/go-lib/c-log"
	"github.com/SealinGp/go-socks5"
	"github.com/SealinGp/wh/config"
	"github.com/SealinGp/wh/pkg/proxy"
	http_svr "github.com/SealinGp/wh/pkg/proxy/http-svr"
)

type App struct {
	ProxyServes []proxy.ProxyServer
	CfgOpt      *config.ConfigOption
}

var (
	_version_ string
	c         string
	e         string
)

func main() {
	flag.StringVar(&c, "c", "config/config.yml", "config path")
	flag.StringVar(&e, "e", "dev", "environment dev|prod")
	flag.Parse()

	app := &App{
		ProxyServes: make([]proxy.ProxyServer, 0, proxy.DefaultMaxServers),
	}

	//config init
	cfgOpt := config.NewConfigOption(c)
	err := cfgOpt.Init()
	if err != nil {
		c_log.E("config init failed. err:%v", err)
		return
	}
	app.CfgOpt = cfgOpt

	logLevel := c_log.LEVEL_ERR

	//dev environment use stderr to log
	if e == "dev" {
		app.CfgOpt.LogPath = ""
		logLevel = c_log.LEVEL_INFO
	}

	//log init(rotate & log level)
	c_log.CLogInit(&c_log.CLogOptions{
		Flag:     log.Lshortfile | log.Ltime,
		Path:     cfgOpt.LogPath,
		LogLevel: logLevel,
	})

	//start proxy
	err = app.StartAll(cfgOpt)
	if err != nil {
		c_log.E("proxy start failed. err:%v", err)
		return
	}
	defer app.CloseAll()

	c_log.I("wh start success. version:%v, pid:%v", _version_, os.Getpid())

	//wait signal for release resources
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	sig := <-sigCh
	c_log.I("wh exit. signal:%v", sig)
}

func (app *App) CloseAll() {
	for _, proxyServer := range app.ProxyServes {
		err := proxyServer.Close()
		if err != nil {
			c_log.E("proxy server close failed. addr:%v, type:%v, err:%v", proxyServer.GetAddr(), proxyServer.GetType(), err)
		}
	}
}

func (app *App) StartAll(opt *config.ConfigOption) error {
	//http server proxy
	for _, httpProxyAddr := range opt.HttpProxyAddrs {
		_, _, err := net.SplitHostPort(httpProxyAddr)
		if err != nil {
			c_log.E("unvalid http proxy addr. addr:%v, err:%v", httpProxyAddr, err)
			continue
		}

		httpProxyServer := http_svr.NewHttpServer(&http_svr.HttpServerOpt{
			Addr: httpProxyAddr,
		})
		if err := httpProxyServer.Start(); err != nil {
			c_log.E("http(s) start failed. addr:%v, err:%s", httpProxyAddr, err)
			return err
		}

		c_log.I("http proxy server start success. addr:%v", httpProxyAddr)
		app.ProxyServes = append(app.ProxyServes, httpProxyServer)
	}

	//socks5 proxy
	for _, socks5ProxyAddr := range opt.SocksProxyAddrs {
		_, _, err := net.SplitHostPort(socks5ProxyAddr)
		if err != nil {
			c_log.E("unvalid socks5 proxy addr ignored. addr:%v, err:%v", socks5ProxyAddr, err)
			continue
		}

		//staticCri := make(socks5.StaticCredentials)
		//staticCri[app.CfgOpt.User] = app.CfgOpt.Pass
		cfg := &socks5.Config{
			//Credentials: staticCri,
			Logger: &LoggerWrap{},
		}
		socks5Server, err := socks5.New(cfg)
		if err != nil {
			c_log.E("new socks5 failed. err:%v", err)
			return err
		}

		socks5Wrap := &Socks5Wrap{
			addr:   socks5ProxyAddr,
			Server: socks5Server,
		}

		if err := socks5Wrap.Start(); err != nil {
			c_log.E("socks5 proxy server start failed. addr:%v, err:%s", socks5ProxyAddr, err)
			return err
		}

		c_log.I("socks5 proxy server start success. addr:%v", socks5ProxyAddr)
		app.ProxyServes = append(app.ProxyServes, socks5Wrap)
	}

	if len(app.ProxyServes) == 0 {
		return errors.New("no proxy servers")
	}

	return nil
}

type LoggerWrap struct {
}

func (loggerWrap *LoggerWrap) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

type Socks5Wrap struct {
	addr string
	*socks5.Server
}

func (socks5Wrap *Socks5Wrap) Start() error {
	go func() {
		socks5Wrap.ListenAndServe("tcp", socks5Wrap.addr)
	}()
	return nil
}

func (socks5Wrap *Socks5Wrap) GetType() string {
	return proxy.SOCK5
}

func (socks5Wrap *Socks5Wrap) GetAddr() string {
	return socks5Wrap.addr
}

func (socks5Wrap *Socks5Wrap) Close() error {
	return nil
}
