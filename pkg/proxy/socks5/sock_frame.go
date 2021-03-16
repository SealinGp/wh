package socks5

//https://www.ietf.org/rfc/rfc1928.txt

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

const (
	DEFAULT_VERSION = 0x05

	//客户端支持的认证方式
	NMETHODS_NONE       = 0x00 //不需要认证
	NMETHODS_GSSAPI     = 0x01 //GSSAPI
	NMETHODS_USERPASS   = 0x02 //用户名密码认证
	NMETHODS_IANA       = 0x03 //保留
	NMETHODS_NOTSUPPORT = 0xff //没有可以接收的认证方式

	//客户端要连接的目标地址的传输层协议
	CMD_TCP  = 0x01
	CMD_BIND = 0x02 //多用于FTP协议
	CMD_UDP  = 0x03

	//客户端连接的目标地址的类型(ipv4,domain,ipv6)
	AYTP_IPV4   = 0x01
	AYTP_DOMAIN = 0x03
	AYTP_IPV6   = 0x04

	AUTH_STATUS_SUCCESS = 0x00
	AUTH_STATUS_FAILED  = 0x01

	REP_SUCCEED                           = 0x00
	REP_GENERAL_SOCKS_SERVER_FAILURE      = 0x01
	REP_CONNECTION_NOT_ALLOWED_BY_RULESET = 0x02
	REP_NETWORK_UNREACHABLE               = 0x03
	REP_HOST_UNREACHABLE                  = 0x04
	REP_CONNECTION_REFUSED                = 0x05
	REP_TTL_EXPIRED                       = 0x06
	REP_COMMAND_NOT_SUPPORTED             = 0x07
	REP_ADDR_TYPE_NOT_SUPPORTED           = 0x08
	REP_UNASSIGNED                        = 0x09

	RSV_DEFAULT = 0x00
)

var (
	AUTH_MAP = map[byte]string{
		NMETHODS_NONE:       "NONE",
		NMETHODS_USERPASS:   "USERPASS",
		NMETHODS_GSSAPI:     "GSSAPI",
		NMETHODS_IANA:       "IANA",
		NMETHODS_NOTSUPPORT: "NOTSUPPORT",
	}
	CMD_MAP = map[byte]string{
		CMD_TCP:  "tcp",
		CMD_BIND: "ftp",
		CMD_UDP:  "udp",
	}
)

type SockAuthRequest struct {
	user     string
	password string
}

func (sockAuthRequest *SockAuthRequest) GetUser() string {
	return sockAuthRequest.user
}
func (sockAuthRequest *SockAuthRequest) GetPassWord() string {
	return sockAuthRequest.password
}

type DST struct {
	ADDR string
	PORT uint16
}

func (dst *DST) String() string {
	return fmt.Sprintf("%v:%v", dst.ADDR, dst.PORT)
}

type SockFrame struct {
	Ver      byte
	Cmd      byte
	Nmethods byte
	Methods  []byte
	Method   byte
	Rsv      uint8
	Aytp     uint8
	Dst      *DST
	Frag     uint8
	data     []byte
}

func NewSockFrame() *SockFrame {
	return &SockFrame{}
}

/**
1.client -> server
[5 3 0 1 2]
客户端请求握手协议帧
+----+----------+----------+
|VER | NMETHODS | METHODS  |
+----+----------+----------+
| 1  |    1     | 1 to 255 |
+----+----------+----------+
*/
func (sockFrame *SockFrame) ReadHandShake(reader io.Reader, ctx context.Context) error {
	dataCh := make(chan []byte, 1)

	go func() {
		data := make([]byte, 1500)
		n, err := reader.Read(data)

		if err != nil {
			log.Printf("[E] read failed. err:%v", err)
			return
		}

		dataCh <- data[:n]
	}()

	var frameData []byte
	select {
	case <-ctx.Done():
		return ctx.Err()
	case frameData = <-dataCh:
		log.Printf("[D] 1.hanshake read data:%v", frameData)
	}

	if len(frameData) < 3 {
		return errors.New("invalid frame")
	}

	if frameData[0] != DEFAULT_VERSION {
		return errors.New(fmt.Sprintf("invalid version %v", frameData[0]))
	}
	sockFrame.Ver = DEFAULT_VERSION

	//鉴权方式
	switch frameData[1] {
	case NMETHODS_USERPASS, NMETHODS_NONE, NMETHODS_IANA, NMETHODS_GSSAPI:
		sockFrame.Nmethods = frameData[1]
	default:
		return errors.New(fmt.Sprintf("invalid Nmethods %v.", frameData))
	}

	//methods
	sockFrame.Methods = frameData[2:]
	return nil
}

/**
1.server -> client
服务端同意握手协议帧
+----+--------+
|VER | METHOD |
+----+--------+
| 1  |   1    |
+----+--------+
*/
func (sockFrame *SockFrame) WriteHandShake(writer io.Writer) (int, error) {
	data := make([]byte, 0, 1500)
	data = append(data, sockFrame.Ver)
	data = append(data, sockFrame.Method)
	return writer.Write(data)
}

/**
ref:https://www.ietf.org/rfc/rfc1929.txt
2.客户端认证请求协议帧(byte unit)
-----+------+------+------+--------+
|VER | ULEN | UNAME| PLEN | PASSWD |
+----+------+------+------+--------+
| 1  |  1   | 1-255|  1   | 1-255  |
+----+------+------+------+--------+
*/
func (sockFrame *SockFrame) ReadAuthReq(ctx context.Context, reader io.Reader) (*SockAuthRequest, error) {
	dataCh := make(chan []byte, 1)
	go func() {
		data := make([]byte, 1500)
		n, err := reader.Read(data)

		if err != nil {
			log.Printf("[E] read failed. err:%v", err)
			return
		}

		dataCh <- data[:n]
		close(dataCh)
	}()

	var frameData []byte
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case frameData = <-dataCh:
		log.Printf("[D] ReadAuth read data:%v", frameData)
	}

	if len(frameData) < 2 {
		return nil, errors.New("invalid frame")
	}

	if frameData[0] != DEFAULT_VERSION {
		return nil, errors.New(fmt.Sprintf("unsupported version %v", frameData[0]))
	}

	sockAuthRequest := &SockAuthRequest{}

	userLen := int(frameData[1])
	if userLen <= 0 || len(frameData) < userLen+1 {
		log.Printf("[E] parse user failed. userLen:%v", userLen)
		return sockAuthRequest, nil
	}
	sockAuthRequest.user = string(frameData[2 : userLen+1])

	passLenIndex := userLen + 2
	if len(frameData)-1 < passLenIndex {
		log.Printf("[E] parse pass index failed. passLenIndex:%v", passLenIndex)
		return sockAuthRequest, nil
	}

	passLen := int(frameData[passLenIndex])
	if len(frameData) < passLen+passLenIndex {
		log.Printf("[E] parse pass failed. passLen:%v", passLen)
		return sockAuthRequest, nil
	}
	sockAuthRequest.password = string(frameData[passLenIndex+1 : passLen+passLenIndex])

	return sockAuthRequest, nil
}

/**
2.服务端返回认证结果协议帧(byte unit)
-----+------+
|VER |STATUS|
+----+------+
| 1  |  1   |
+----+------+
*/
func (sockFrame *SockFrame) WriteAuthResp(writer io.Writer, isSuccess bool) (int, error) {
	var status byte
	status = AUTH_STATUS_SUCCESS
	if !isSuccess {
		status = AUTH_STATUS_FAILED
	}

	data := make([]byte, 0, 2)
	data = append(data, DEFAULT_VERSION)
	data = append(data, status)
	return writer.Write(data)
}

/**
client -> server
3.客户端指定目标传输协议(tcp|udp),地址,端口
[5 1 0 1 127 0 0 1 4 211]
+----+-----+-------+------+----------+----------+
|VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
+----+-----+-------+------+----------+----------+
| 1  |  1  | X'00' |  1   | Variable |    2     |
+----+-----+-------+------+----------+----------+
*/
func (sockFrame *SockFrame) ReadInstruction(ctx context.Context, reader io.Reader) error {
	dataCh := make(chan []byte, 1)

	go func() {
		data := make([]byte, 150)
		n, err := reader.Read(data)
		if err != nil {
			log.Printf("[E] read failed. err:%v", err)
			return
		}

		dataCh <- data[:n]
	}()

	var frameData []byte
	select {
	case <-ctx.Done():
		return ctx.Err()
	case frameData = <-dataCh:
		log.Printf("[D] 2.read instruction data:%v", frameData)
	}

	if len(frameData) < 6 {
		return errors.New("invalid frame")
	}

	//version
	if frameData[0] != DEFAULT_VERSION {
		return errors.New(fmt.Sprintf("unsupported sock version %v", frameData[0]))
	}
	sockFrame.Ver = frameData[0]

	switch frameData[1] {
	case CMD_TCP, CMD_UDP, CMD_BIND:
		sockFrame.Cmd = frameData[1]
	default:
		return errors.New(fmt.Sprintf("unsupported proxy protocol %v", frameData[1]))
	}

	sockFrame.Rsv = frameData[2]
	sockFrame.Dst = &DST{}
	switch frameData[3] {
	case AYTP_IPV4:
		if len(frameData) <= 4+net.IPv4len {
			return errors.New("ipv4 parse failed")
		}
		var a byte
		var b byte
		var c byte
		var d byte
		for i, x := range frameData[4 : 4+net.IPv4len] {
			if i == 0 {
				a = x
			}
			if i == 1 {
				b = x
			}
			if i == 2 {
				c = x
			}
			if i == 3 {
				d = x
			}
		}
		ip := net.IPv4(a, b, c, d)

		sockFrame.Aytp = AYTP_IPV4
		sockFrame.Dst.ADDR = ip.To4().String()
		sockFrame.Dst.PORT = binary.BigEndian.Uint16(frameData[4+net.IPv4len:])
	case AYTP_IPV6:
		if len(frameData) <= 4+net.IPv6len {
			return errors.New("ipv6 parse failed")
		}
		ip := net.ParseIP(string(frameData[4 : 4+net.IPv6len]))

		sockFrame.Aytp = AYTP_IPV6
		sockFrame.Dst.ADDR = ip.To16().String()
		sockFrame.Dst.PORT = binary.BigEndian.Uint16(frameData[4+net.IPv6len:])
	case AYTP_DOMAIN:
		dstAddrLen := int(frameData[4])
		if len(frameData) < 4+dstAddrLen {
			return errors.New("domain parse failed")
		}

		sockFrame.Aytp = AYTP_DOMAIN
		sockFrame.Dst.ADDR = string(frameData[4 : 4+dstAddrLen])
		sockFrame.Dst.PORT = binary.BigEndian.Uint16(frameData[4+dstAddrLen:])
	default:
		return errors.New("unsupported AYTP type")
	}

	return nil
}

/**
server -> client
3.服务端返回目标传输协议,地址,端口的代理连接情况
+----+-----+-------+------+----------+----------+
|VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
+----+-----+-------+------+----------+----------+
| 1  |  1  | X'00' |  1   | Variable |    2     |
+----+-----+-------+------+----------+----------+
*/

func (sockFrame *SockFrame) WriteInstruction(writer io.Writer, connFunc func() (ipv4 []byte, port []byte, err error)) (int, error) {
	rep := REP_SUCCEED

	ipv4, port, connErr := connFunc()
	if connErr != nil {
		rep = REP_GENERAL_SOCKS_SERVER_FAILURE
	}

	data := make([]byte, 0, 4+len(ipv4)+len(port))
	data = append(data, DEFAULT_VERSION)
	data = append(data, byte(rep))
	data = append(data, RSV_DEFAULT)
	data = append(data, AYTP_IPV4)
	data = append(data, ipv4...)
	data = append(data, port...)

	return writer.Write(data)
}

/**
FRAG:Current fragment number
udp传输的数据帧
+------+------+----------+----------+---------------+
|RSV | FRAG | ATYP | DST.ADDR | DST.PORT |   DATA   |
+----+------+------+----------+----------+----------+
| 2  |  1   |  1   | Variable |    2     | Variable |
+----+------+------+----------+----------+----------+
*/
func (sockFrame *SockFrame) ReadUdpData(reader io.Reader, ctx context.Context) error {
	dataCh := make(chan []byte, 1)

	go func() {
		data := make([]byte, 150)
		n, err := reader.Read(data)
		if err != nil {
			log.Printf("[E] read failed. err:%v", err)
			return
		}

		dataCh <- data[:n]
	}()

	var frameData []byte
	select {
	case <-ctx.Done():
		return ctx.Err()
	case frameData = <-dataCh:
		log.Printf("[D] 2.read instruction data:%v", frameData)
	}

	if len(frameData) <= 7 {
		return errors.New("invalid frame")
	}

	sockFrame.Frag = frameData[2]
	sockFrame.Dst = &DST{}
	switch frameData[3] {
	case AYTP_IPV4:
		if len(frameData) <= 4+net.IPv4len {
			return errors.New("ipv4 parse failed")
		}
		var a byte
		var b byte
		var c byte
		var d byte
		for i, x := range frameData[4 : 4+net.IPv4len] {
			if i == 0 {
				a = x
			}
			if i == 1 {
				b = x
			}
			if i == 2 {
				c = x
			}
			if i == 3 {
				d = x
			}
		}
		ip := net.IPv4(a, b, c, d)

		sockFrame.Aytp = AYTP_IPV4
		sockFrame.Dst.ADDR = ip.To4().String()
		sockFrame.Dst.PORT = binary.BigEndian.Uint16(frameData[4+net.IPv4len : 5+net.IPv4len])
		sockFrame.data = frameData[5+net.IPv4len:]
	case AYTP_IPV6:
		if len(frameData) <= 4+net.IPv6len {
			return errors.New("ipv6 parse failed")
		}
		ip := net.ParseIP(string(frameData[4 : 4+net.IPv6len]))

		sockFrame.Aytp = AYTP_IPV6
		sockFrame.Dst.ADDR = ip.To16().String()
		sockFrame.Dst.PORT = binary.BigEndian.Uint16(frameData[4+net.IPv6len : 5+net.IPv6len])
		sockFrame.data = frameData[5+net.IPv6len:]
	case AYTP_DOMAIN:
		dstAddrLen := int(frameData[4])
		if len(frameData) < 4+dstAddrLen {
			return errors.New("domain parse failed")
		}

		sockFrame.Aytp = AYTP_DOMAIN
		sockFrame.Dst.ADDR = string(frameData[4 : 4+dstAddrLen])
		sockFrame.Dst.PORT = binary.BigEndian.Uint16(frameData[4+dstAddrLen : 5+dstAddrLen])
		sockFrame.data = frameData[5+dstAddrLen:]
	}

	return errors.New("unsupported AYTP type")
}
