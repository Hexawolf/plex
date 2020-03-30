// plex - User-space UDP broker to replace IP multicasts
// Copyright (C) 2020 Hexawolf <hexawolf@hexanet.dev>
// See LICENSE file for more info
package plex

import (
	"io"
	"log"
	"net"
	"os"
	"sync"
)

type Plex struct {
	log    *log.Logger

	// subscribers registry
	subs   map[string]io.WriteCloser
	// publishers registry
	pubs   map[string]io.ReadCloser
	// registry access mutex
	regM   sync.Mutex

	// anyone can write
	in     *io.PipeWriter
	// to maintain copy-on-write instead of go's sequential reading,
	// only one thread is allowed to read at a given time
	out    *io.PipeReader
}
const UDP = "udp"

// NewPlex creates a new UDP multicast instance, binding a listener to supplied local address (_laddr)
// and allocating buffer with given size (_bsize).
func NewPlex(bsize uint16, logger *log.Logger) (mp Plex, err error) {
	if logger == nil { logger = log.New(os.Stdout, "plex ", logger.Flags()) }
	mp.log = logger
	mp.subs = make(map[string]io.WriteCloser)
	mp.pubs = make(map[string]io.ReadCloser)
	mp.out, mp.in = io.Pipe()
	go mp.plex(bsize)
	return
}

func (mp *Plex) plex(_bsize uint16) { for {
	buf := make([]byte, _bsize)
	_, err := mp.out.Read(buf)
	if err != nil { mp.log.Println(err); return }
	for k, s := range mp.subs {
		if s == nil { continue }
		_, err := s.Write(buf)
		if err != nil {
			mp.log.Println(err)
			mp.Unsubscribe(k)
		}
	}
}}

func (mp *Plex) exists(name string) (sub bool, pub bool) {
	if _, ok := mp.subs[name]; ok {
		sub = true
	}
	if _, ok := mp.pubs[name]; ok {
		pub = true
	}
	return
}

func (mp *Plex) Exists(name string) (sub bool, pub bool) {
	mp.regM.Lock()
	defer mp.regM.Unlock()
	return mp.exists(name)
}

func (mp *Plex) Unsubscribe(name string) {
	mp.regM.Lock()
	defer mp.regM.Unlock()
	sub, pub := mp.exists(name);
	if sub {
		mp.subs[name].Close()
		delete(mp.subs, name)
	}
	if pub {
		mp.pubs[name].Close()
		delete(mp.pubs, name)
	}
}

func (mp *Plex) inSafe(pub io.Reader) {
	_, err := io.Copy(mp.in, pub)
	if err != nil {
		mp.log.Println(err)
	}
}

func (mp *Plex) Subscribe(name string, sub io.WriteCloser, pub io.ReadCloser) {
	mp.regM.Lock()
	defer mp.regM.Unlock()
	if sub != nil {
		mp.subs[name] = sub
	}
	if pub != nil {
		go mp.inSafe(pub)
	}
}

func (mp *Plex) ListenUDP(laddr string) (err error) {
	var _laddr *net.UDPAddr
	_laddr, err = net.ResolveUDPAddr(UDP, laddr)
	if err != nil { return }
	var sock *net.UDPConn
	sock, err = net.ListenUDP(UDP, _laddr)
	if err != nil { return }
	mp.log.Println("Listening on", sock.LocalAddr().String());
	_, err = io.Copy(mp.in, sock)
	return
}

func (mp *Plex) SubscribeUDP(_raddr string) (err error) {
	var raddr *net.UDPAddr
	var conn *net.UDPConn
	raddr, err = net.ResolveUDPAddr(UDP, _raddr)
	if err != nil { return }
	conn, err = net.DialUDP(UDP, nil, raddr)
	if err != nil { return }
	mp.Subscribe(_raddr, conn, conn)
	mp.log.Println("Subscribed:", _raddr)
	return
}

func (mp *Plex) Close() (err error) {
	mp.regM.Lock()
	defer mp.regM.Unlock()
	for _, v := range mp.pubs { v.Close() }
	mp.in.Close(); mp.out.Close()
	for _, v := range mp.subs { v.Close() }
	return
}
