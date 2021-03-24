package socks5

import (
	"bytes"
	"context"
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
}
