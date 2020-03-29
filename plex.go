// plex - User-space UDP broker to replace IP multicasts
// Copyright (C) Hexawolf <hexawolf@hexanet.dev>
// See LICENSE file for more info
package plex

import (
	"io"
	"log"
	"net"
	"os"
)

type Plex struct {
	log    *log.Logger
	sock   *net.UDPConn
	in     *io.PipeWriter
	out    *io.PipeReader
	subs   map[string]io.WriteCloser
}
const UDP = "udp"

// NewPlex creates a new UDP multicast instance, binding a listener to supplied local address (_laddr)
// and allocating buffer with given size (_bsize).
func NewPlex(_laddr string, _bsize uint16, _log *log.Logger) (mp Plex, err error) {
	if _log == nil { _log = log.New(os.Stdout, "plex ", log.Flags()) }
	mp.log = _log
	var laddr *net.UDPAddr
	laddr, err = net.ResolveUDPAddr(UDP, _laddr)
	if err != nil { return }
	mp.sock, err = net.ListenUDP(UDP, laddr)
	if err != nil { return }
	mp.subs = make(map[string]io.WriteCloser)
	mp.out, mp.in = io.Pipe()
	go func() { for {
		buf := make([]byte, _bsize)
		_, err := mp.out.Read(buf)
		if err != nil { mp.log.Println(err); return }
		for k, s := range mp.subs {
			if s == nil { continue }
			_, err := s.Write(buf)
			if err != nil {
				mp.log.Println(err)
				delete(mp.subs, k)
			}
		}
	}}()
	return
}

func (mp *Plex) ListenUDP() (err error) {
	mp.log.Println("Listening on", mp.sock.LocalAddr().String());
	_, err = io.Copy(mp.in, mp.sock)
	return
}

func (mp *Plex) SubscribeUDP(_raddr string, out *net.UDPConn) (err error) {
	var raddr *net.UDPAddr
	raddr, err = net.ResolveUDPAddr(UDP, _raddr)
	if err != nil { return }
	out, err = net.DialUDP(UDP, nil, raddr)
	if err != nil { return }
	mp.subs[_raddr] = out
	mp.log.Println("Subscribed:", _raddr)
	return
}

func (mp *Plex) SubscribeAllUDP() {
	
}

func (mp *Plex) Close() (err error) {
	err = mp.sock.Close()
	mp.in.Close(); mp.out.Close()
	for _, v := range mp.subs { v.Close() }
	return
}
