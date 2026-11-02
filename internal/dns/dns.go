// Package dns is a tiny authoritative-for-a-list, forwarding-for-the-rest DNS server.
package dns

import (
	"bufio"
	_ "embed"
	"io"
	"log"
	"net"
	"net/netip"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

//go:embed hosts.txt
var DefaultHosts string

// ParseHosts reads hosts-style lines into a lower-case name -> address map.
func ParseHosts(r io.Reader) map[string][4]byte {
	m := map[string][4]byte{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line, _, _ := strings.Cut(sc.Text(), "#")
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		ip, err := netip.ParseAddr(f[0])
		if err != nil || !ip.Is4() {
			continue
		}
		for _, n := range f[1:] {
			m[strings.ToLower(n)+"."] = ip.As4()
		}
	}
	return m
}

// LoadHosts reads path, or the compiled-in default when path is empty.
func LoadHosts(path string) (map[string][4]byte, error) {
	if path == "" {
		return ParseHosts(strings.NewReader(DefaultHosts)), nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseHosts(f), nil
}

type Server struct {
	Hosts    map[string][4]byte
	Upstream string // host:port
	TTL      uint32

	mu   sync.Mutex
	seen map[string]time.Time
}

func (s *Server) logOnce(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.seen == nil {
		s.seen = map[string]time.Time{}
	}
	if time.Since(s.seen[name]) > time.Minute {
		s.seen[name] = time.Now()
		log.Printf("answered %s", name)
	}
}

// Reply answers one query packet. network is "udp" or "tcp" and picks the upstream transport.
func (s *Server) Reply(q []byte, network string) []byte {
	var p dnsmessage.Parser
	h, err := p.Start(q)
	if err != nil || h.Response {
		return nil
	}
	qs, err := p.Question()
	if err != nil {
		return nil
	}
	name := strings.ToLower(qs.Name.String())
	ip, ok := s.Hosts[name]
	if !ok || (qs.Class != dnsmessage.ClassINET) || (qs.Type != dnsmessage.TypeA && qs.Type != dnsmessage.TypeAAAA) {
		return s.forward(q, network)
	}
	rh := dnsmessage.Header{ID: h.ID, Response: true, Authoritative: true, RecursionDesired: h.RecursionDesired}
	b := dnsmessage.NewBuilder(nil, rh)
	b.EnableCompression()
	b.StartQuestions()
	b.Question(qs)
	b.StartAnswers()
	if qs.Type == dnsmessage.TypeA { // AAAA: NOERROR, no data
		b.AResource(dnsmessage.ResourceHeader{Name: qs.Name, Class: dnsmessage.ClassINET, TTL: s.TTL}, dnsmessage.AResource{A: ip})
		s.logOnce(name)
	}
	out, err := b.Finish()
	if err != nil {
		return nil
	}
	return out
}

func (s *Server) forward(q []byte, network string) []byte {
	c, err := net.DialTimeout(network, s.Upstream, 3*time.Second)
	if err != nil {
		return servfail(q)
	}
	defer c.Close()
	c.SetDeadline(time.Now().Add(4 * time.Second))
	buf := make([]byte, 0, 2+len(q))
	if network == "tcp" {
		buf = append(buf, byte(len(q)>>8), byte(len(q)))
	}
	if _, err := c.Write(append(buf, q...)); err != nil {
		return servfail(q)
	}
	if network == "tcp" {
		var l [2]byte
		if _, err := io.ReadFull(c, l[:]); err != nil {
			return servfail(q)
		}
		r := make([]byte, int(l[0])<<8|int(l[1]))
		if _, err := io.ReadFull(c, r); err != nil {
			return servfail(q)
		}
		return r
	}
	r := make([]byte, 4096)
	n, err := c.Read(r)
	if err != nil {
		return servfail(q)
	}
	return r[:n]
}

func servfail(q []byte) []byte {
	if len(q) < 12 {
		return nil
	}
	r := append([]byte(nil), q[:12]...)
	r[2] = r[2]&0x01 | 0x80
	r[3] = 2
	r[6], r[7], r[8], r[9], r[10], r[11] = 0, 0, 0, 0, 0, 0
	return r
}

// Serve runs UDP and TCP on addr until either fails.
func (s *Server) Serve(addr string) error {
	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.serve(pc, ln)
}

func (s *Server) serve(pc net.PacketConn, ln net.Listener) error {
	errc := make(chan error, 2)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, from, err := pc.ReadFrom(buf)
			if err != nil {
				errc <- err
				return
			}
			q := append([]byte(nil), buf[:n]...)
			go func() {
				if r := s.Reply(q, "udp"); r != nil {
					pc.WriteTo(r, from)
				}
			}()
		}
	}()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				errc <- err
				return
			}
			go s.tcpConn(c)
		}
	}()
	return <-errc
}

func (s *Server) tcpConn(c net.Conn) {
	defer c.Close()
	for {
		c.SetDeadline(time.Now().Add(10 * time.Second))
		var l [2]byte
		if _, err := io.ReadFull(c, l[:]); err != nil {
			return
		}
		q := make([]byte, int(l[0])<<8|int(l[1]))
		if _, err := io.ReadFull(c, q); err != nil {
			return
		}
		r := s.Reply(q, "tcp")
		if r == nil {
			return
		}
		if _, err := c.Write(append([]byte{byte(len(r) >> 8), byte(len(r))}, r...)); err != nil {
			return
		}
	}
}
