package main

import (
	"log"

	"github.com/SealinGp/go-socks5"
)

func main() {

	c := &socks5.Config{}
	server, err := socks5.New(c)
	if err != nil {
		log.Printf("[E] new failed. err:%v", err)
		return
	}

	err = server.ListenAndServe("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Printf("[E] list failed. err:%v", err)
	}
}
