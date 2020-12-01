package main

import (
	"flag"
	"github.com/SealinGp/wh/pkg/proxy"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	httpProxy := proxy.NewHttpPxy(&proxy.HttpPxyOpt{
		Debug: true,
	})
	_ = httpProxy.Start()

	sigCh := make(chan os.Signal)
	closeDoneCh := make(chan bool)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		sig := <-sigCh
		log.Printf("[I] received sig closing...sig:%s \n", sig)

		//todo closing ...
		err := httpProxy.Close()
		if err != nil {
			log.Printf("[E] close failed. err:%s", err)
		}

		close(closeDoneCh)
		return
	}()
	<-closeDoneCh
}
