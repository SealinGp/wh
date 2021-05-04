package socks5

import (
	"bytes"
	"context"
	"encoding/binary"
	"log"
	"testing"
)

func TestSockFrame_ReadInstruction(t *testing.T) {
	domainData := []byte{5, 1, 0, 3, 14, 119, 119, 119, 46, 103, 111, 111, 103, 108, 101, 46, 99, 111, 109, 1, 187}

	sf := NewSockFrame()
	sf.ReadInstruction(context.Background(), bytes.NewReader(domainData))
	log.Printf("%s", sf.Dst.String())

	ipv4Data := []byte{5, 1, 0, 1, 127, 0, 0, 1, 4, 211}
	sf.ReadInstruction(context.Background(), bytes.NewReader(ipv4Data))
	log.Printf("%s", sf.Dst.String())

	handshakData := []byte{5, 1, 0}
	log.Printf("[D] handshakData1:%v", handshakData[1])

	portData := make([]byte, 2)
	port := []byte("80")
	binary.BigEndian.PutUint16(portData, binary.BigEndian.Uint16(port))
	log.Printf("[D] portData:%v", portData)
}
