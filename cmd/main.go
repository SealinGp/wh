package main

import (
	"context"
	"flag"
	"github.com/SealinGp/wh/pkg/transport/tcp"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	httpProxy := tcp.NewServer(&tcp.ServerOpt{
		Addr:      *addr,
		ProxyType: tcp.HTTP_PROXY,
	})
	if err := httpProxy.Start(); err != nil {
		log.Printf("[E] http start failed. err:%s\n", err)
		return
	}
	log.Printf("[D] tcp listening in %s...", *addr)

	//release resources
	sigCh := make(chan os.Signal)
	closeDoneCh := make(chan bool)
	go func() {
		log.Printf("[I] received sig closing...sig:%s \n", <-sigCh)

		//todo closing ...
		err := httpProxy.Close()
		if err != nil {
			log.Printf("[E] close failed. err:%s", err)
		}

		close(closeDoneCh)
		return
	}()
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGHUP)
	<-closeDoneCh

	//waiting for release
	log.Printf("[I] wating 3s to release all resources...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	<-ctx.Done()
	log.Printf("[I] release done.")
}
