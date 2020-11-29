package main

import (
	"flag"
	http_proxy "github.com/SealinGp/wh/http-proxy"
	"github.com/SealinGp/wh/pkg/proxy"
	"log"
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
	httpProxy := proxy.NewHttpProxy(&proxy.HttpProxyOpt{
		Address: *address,
	})
	err := httpProxy.Start()
	defer httpProxy.Close()
	if err != nil {
		log.Printf("[E] start failed. err:%s", err)
		return
	}
}
