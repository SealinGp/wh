package main

import (
	"flag"
	"github.com/SealinGp/wh/pkg/proxy"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	addr = flag.String("addr", "", "listen address")
	help = flag.Bool("help", false, "print help")
)

func main() {
	flag.Parse()
	if *addr == "" || *help {
		flag.PrintDefaults()
		return
	}

	httpProxy := proxy.NewHttpServer(&proxy.HttpServerOpt{
		Addr: *addr,
	})
	if err := httpProxy.Start(); err != nil {
		log.Printf("[E] http start failed. err:%s\n", err)
		return
	}
	log.Printf("[D] tcp listening in %s...", *addr)

	//release resources
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGHUP)
	<-sigCh

	err := httpProxy.Close()
	if err != nil {
		log.Printf("[E] close failed. err:%v", err)
	}
}
