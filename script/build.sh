#!/bin/bash
#1.todo xargs
#2.
#2020/12/02 22:06:39 [I] listening in :1234 ...
#2020/12/02 22:07:46 received req. url://www.google.com:443 headers:map[Proxy-Connection:[keep-alive] User-Agent:[Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36]], method:CONNECT,clientAddr:218.18.239.38:54073
#2020/12/02 22:07:46 [E] get url failed. err:EOF

find $PWD -type f -name main.go
go build -o $PWD/wh main.go
