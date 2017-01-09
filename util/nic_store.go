package util

import (
	"sync"

	"github.com/vishvananda/netlink"
)

type nicStore struct {
	store map[string]netlink.Link
	lock  sync.Mutex
}

var NicStore = &nicStore{
	store: make(map[string]netlink.Link),
}

func (s *nicStore) Add(ip string, nic netlink.Link) {
	s.lock.Lock()
	s.store[ip] = nic
	s.lock.Unlock()
}

func (s *nicStore) Delete(ip string) netlink.Link {
	s.lock.Lock()
	link := s.store[ip]
	delete(s.store, ip)
	s.lock.Unlock()
	return link
}
