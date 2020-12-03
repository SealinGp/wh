package main

import (
	"context"
	"flag"
	tcp "github.com/SealinGp/wh/pkg/proxy/tcp"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/**
https:
CONNECT www.google.com:443 HTTP/1.1
Host: www.google.com:443
Proxy-Connection: keep-alive
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.95 Safari/537.36

http:
GET http://www.flysnow.org/ HTTP/1.1
Host: www.flysnow.org
Proxy-Connection: keep-alive
Upgrade-Insecure-Requests: 1
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.95 Safari/537.36



server
1.接受连接
2.转发连接给Host
*/
var (
	address = flag.String("addr", "", "listen address")
)

func main() {
	flag.Parse()
	if *address == "" {
		log.Println("-address required")
		return
	}

	httpProxy := tcp.NewServer(&tcp.ServerOpt{
		Addr:      *address,
		ProxyType: tcp.HTTP_PROXY,
	})
	if err := httpProxy.Start(); err != nil {
		log.Printf("[E] http start failed. err:%s\n", err)
		return
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	<-ctx.Done()
}
