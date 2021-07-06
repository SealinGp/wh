package svrs

import (
	"errors"
	"io"
	"sync"
)

var (
	s *servers
)

func Init() {
	s = &servers{
		svrs: make(map[string]*serverItem),
	}
}

func Add(name, addr string, server io.Closer) error {
	s.Lock()
	defer s.Unlock()

	_, ok := s.svrs[addr]
	if ok {
		return errors.New("addr exists")
	}

	si := &serverItem{
		serverElement: serverElement{
			name: name,
			addr: addr,
		},
		Closer: server,
	}

	s.svrs[addr] = si
	return nil
}

func GetServerElements() []*serverElement {
	s.RLock()
	defer s.RUnlock()

	ses := make([]*serverElement, 0, len(s.svrs))
	for _, si := range s.svrs {
		ses = append(ses, &serverElement{
			name: si.name,
			addr: si.addr,
		})
	}

	return ses
}

func Del(addr string) {
	s.Lock()
	defer s.Unlock()

	svr, ok := s.svrs[addr]
	if !ok {
		return
	}

	delete(s.svrs, addr)
	svr.Close()
}

func CloseAll() {
	s.Lock()
	defer s.Unlock()

	for _, c := range s.svrs {
		c.Close()
	}
}

type servers struct {
	svrs map[string]*serverItem
	sync.RWMutex
}

type serverElement struct {
	name string
	addr string
}

type serverItem struct {
	serverElement
	io.Closer
}
