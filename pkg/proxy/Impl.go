package proxy

import "io"

type ProxyServer interface {
	Start() error
	GetType() string
	GetAddr() string
	io.Closer
}
