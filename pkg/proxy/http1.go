package proxy

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type HttpPxy struct {
	server *http.Server
	debug  bool
}

type HttpPxyOpt struct {
	Debug bool
	Addr  string
}

func NewHttpPxy(opt *HttpPxyOpt) *HttpPxy {
	httpProxy := &HttpPxy{
		debug: opt.Debug,
		server: &http.Server{
			Addr: opt.Addr,
		},
	}
	return httpProxy
}

func (httpPxy *HttpPxy) Start() error {
	mux := http.NewServeMux()
	mux.Handle("/", httpPxy)
	httpPxy.server.Handler = http.NewServeMux()

	go func() {
		log.Printf("[I] listening in %s ...", httpPxy.server.Addr)
		if err := httpPxy.server.ListenAndServe(); err != nil {
			log.Printf("[E] listenAndServe failed. err:%s", err)
		}
	}()
	return nil
}

func (httpPxy *HttpPxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Printf("Received request %s %s %s\n", req.Method, req.Host, req.RemoteAddr)

	// step 1
	outReq := new(http.Request)
	*outReq = *req // this only does shallow copies of maps

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if prior, ok := outReq.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		outReq.Header.Set("X-Forwarded-For", clientIP)
	}

	// step 2
	res, err := http.DefaultTransport.RoundTrip(outReq)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	// step 3
	for key, value := range res.Header {
		for _, v := range value {
			rw.Header().Add(key, v)
		}
	}

	rw.WriteHeader(res.StatusCode)
	if _, err := io.Copy(rw, res.Body); err != nil {
		log.Printf("[E] io.Copy failed.err:%s \n", err)
	}
	defer res.Body.Close()
}

func (httpPxy *HttpPxy) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return httpPxy.server.Shutdown(ctx)
}
