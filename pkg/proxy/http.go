package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type HttpPxy struct {
	server *http.Server
	client *http.Client
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
		client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
	httpProxy.server.Handler = httpProxy
	return httpProxy
}

func (httpPxy *HttpPxy) Start() error {
	httpPxy.server.Handler = httpPxy
	go func() {
		log.Printf("[I] listening in %s ...", httpPxy.server.Addr)
		if err := httpPxy.server.ListenAndServe(); err != nil {
			log.Printf("[E] listenAndServe failed. err:%s", err)
		}
	}()
	return nil
}

func (httpPxy *HttpPxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if httpPxy.debug {
		log.Printf("received req. url:%s headers:%v, method:%s,clientAddr:%s", req.URL, req.Header, req.Method, req.RemoteAddr)
	}
	defer req.Body.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	url, err := httpPxy.GetUrl(req)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Printf("[E] get url failed. err:%s\n", err)
		return
	}

	body, err := httpPxy.GetBody(req)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Printf("[E] get url failed. err:%s\n", err)
		return
	}

	request, err := http.NewRequestWithContext(ctx, req.Method, url, body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Printf("[E] new request failed. err:%s\n", err)
		return
	}
	request.Header = req.Header.Clone()

	resp, err := httpPxy.client.Do(request)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Printf("[E] client do failed. err:%s\n", err)
		return
	}

	_, _ = io.Copy(rw, resp.Body)

}
func (httpPxy *HttpPxy) GetUrl(req *http.Request) (string, error) {
	var url, httpVer string

	//get dst url
	connect := req.Header.Get("CONNECT")
	_, err := fmt.Sscanf(connect, "%s%s", &url, &httpVer)
	if err != nil {
		return "", err
	}

	//check http version
	_, _, ok := http.ParseHTTPVersion(httpVer)
	if !ok {
		return "", errors.New("unknown http version")
	}

	//scheme
	scheme := "http://"
	if strings.Contains(url, "443") {
		url = strings.ReplaceAll(url, ":443", "")
		scheme = "https://"
	}
	url = fmt.Sprintf("%s%s", scheme, url)

	return url, nil
}

func (httpPxy *HttpPxy) GetBody(req *http.Request) (io.Reader, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	by := make([]byte, 1500)
	buf := bytes.NewBuffer(by)
	_, err = buf.Write(body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (httpPxy *HttpPxy) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return httpPxy.server.Shutdown(ctx)
}
