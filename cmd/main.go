package main

import (
	"errors"
	http_svr "github.com/SealinGp/wh/pkg/proxy/http-svr"
	"github.com/SealinGp/wh/pkg/proxy/socks5"
	"github.com/jessevdk/go-flags"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type ProxyServer interface {
	Start() error
	GetType() string
	GetAddr() string
	io.Closer
}

type CmdOptions struct {
	HttpProxyAddrs  []string `short:"p" long:"httpProxyAddr" description:"http(s) proxy addresses" required:"no"`
	SocksProxyAddrs []string `short:"s" long:"socksProxyAddrs" description:"socks5 proxy addresses" required:"no"`
}

func (app *App) CloseAll() {
	for _, proxyServer := range app.ProxyServes {
		err := proxyServer.Close()
		if err != nil {
			log.Printf("[E] proxy server close failed. addr:%v, type:%v, err:%v", proxyServer.GetAddr(), proxyServer.GetType(), err)
		}
	}
}

func (app *App) StartAll(opt *CmdOptions) error {
	//http server proxy
	for _, httpProxyAddr := range opt.HttpProxyAddrs {
		_, _, err := net.SplitHostPort(httpProxyAddr)
		if err != nil {
			log.Printf("[E] unvalid http proxy addr. addr:%v, err:%v", httpProxyAddr, err)
			continue
		}

		httpProxyServer := http_svr.NewHttpServer(&http_svr.HttpServerOpt{
			Addr: httpProxyAddr,
		})
		if err := httpProxyServer.Start(); err != nil {
			log.Printf("[E] http(s) start failed. addr:%v, err:%s", httpProxyAddr, err)
			return err
		}

		log.Printf("[I] http proxy server start success. addr:%v", httpProxyAddr)
		app.ProxyServes = append(app.ProxyServes, httpProxyServer)
	}

	//socks5 proxy
	for _, socks5ProxyAddr := range opt.SocksProxyAddrs {
		_, _, err := net.SplitHostPort(socks5ProxyAddr)
		if err != nil {
			log.Printf("[E] unvalid socks5 proxy addr ignored. addr:%v, err:%v", socks5ProxyAddr, err)
			continue
		}

		socks5ProxyServer := socks5.NewSockServer(&socks5.SockServerOpt{
			Addr: socks5ProxyAddr,
		})
		if err := socks5ProxyServer.Start(); err != nil {
			log.Printf("[E] socks5 proxy server start failed. addr:%v, err:%s", socks5ProxyAddr, err)
			return err
		}

		log.Printf("[I] socks5 proxy server start success. addr:%v", socks5ProxyAddr)
		app.ProxyServes = append(app.ProxyServes, socks5ProxyServer)
	}

	if len(app.ProxyServes) == 0 {
		return errors.New("no proxy servers")
	}

	return nil
}

type App struct {
	ProxyServes []ProxyServer
}

const DefaultMaxServers = 20

func main() {
	cmdOptions := &CmdOptions{}
	_, err := flags.Parse(cmdOptions)
	if err != nil {
		return
	}

	app := &App{
		ProxyServes: make([]ProxyServer, 0, DefaultMaxServers),
	}
	defer app.CloseAll()

	err = app.StartAll(cmdOptions)
	if err != nil {
		log.Printf("[E] proxy start failed. err:%v", err)
		return
	}

	//release resources
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGHUP)
	<-sigCh
}
