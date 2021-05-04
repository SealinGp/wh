package http_svr

import "errors"

var (
	ErrNotTunnelProxy = errors.New("not tunnel proxy, conn finished")
)
