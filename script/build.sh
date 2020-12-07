#!/bin/bash
#1.todo xargs
#2.todo read from closed network connection
#2020/12/07 21:56:57 [D] tcp listening in :1234...
#2020/12/07 21:57:04 [I] dst conn established. src:127.0.0.1:50771, dst:113.105.155.219:443
#2020/12/07 21:57:04 [I] dst conn established. src:127.0.0.1:50769, dst:101.37.113.127:443
#2020/12/07 21:57:04 [I] dst conn established. src:127.0.0.1:50770, dst:180.163.150.38:443
#2020/12/07 21:57:07 [E] read from tcpConn failed. err:read tcp 127.0.0.1:1234->127.0.0.1:50770: use of closed network connection
#2020/12/07 21:57:09 [I] dst conn established. src:127.0.0.1:50775, dst:104.16.126.175:443
#2020/12/07 21:57:24 [E] read from tcpConn failed. err:read tcp 127.0.0.1:1234->127.0.0.1:50775: use of closed network connection
#2020/12/07 21:57:34 [E] read from tcpConn failed. err:read tcp 192.168.1.21:50772->101.37.113.127:443: use of closed network connection
#2020/12/07 21:57:34 [E] read from tcpConn failed. err:read tcp 192.168.1.21:50773->113.105.155.219:443: use of closed network connection

#find $PWD -type f -name main.go
addr=":1234"

go build -o $PWD/wh main.go
./wh -addr "${addr}"
