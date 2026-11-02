package dns

import (
	"net"
	"os"
	"strings"
	"testing"

	"golang.org/x/net/dns/dnsmessage"
)

func query(t *testing.T, network, addr, name string, typ dnsmessage.Type) dnsmessage.Message {
	t.Helper()
	n := dnsmessage.MustNewName(name)
	b := dnsmessage.NewBuilder(nil, dnsmessage.Header{ID: 7, RecursionDesired: true})
	b.StartQuestions()
	b.Question(dnsmessage.Question{Name: n, Type: typ, Class: dnsmessage.ClassINET})
	q, _ := b.Finish()
	c, err := net.Dial(network, addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if network == "tcp" {
		q = append([]byte{byte(len(q) >> 8), byte(len(q))}, q...)
	}
	c.Write(q)
	buf := make([]byte, 4096)
	k, _ := c.Read(buf)
	if network == "tcp" {
		buf = buf[2:]
		k -= 2
	}
	var m dnsmessage.Message
	if err := m.Unpack(buf[:k]); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestServe(t *testing.T) {
	// fake upstream: answers every A query with 9.9.9.9
	up, _ := net.ListenPacket("udp", "127.0.0.1:0")
	up2 := &Server{Hosts: map[string][4]byte{"other.example.": {9, 9, 9, 9}}}
	go func() {
		buf := make([]byte, 512)
		for {
			n, from, err := up.ReadFrom(buf)
			if err != nil {
				return
			}
			up.WriteTo(up2.Reply(buf[:n], "udp"), from)
		}
	}()
	s := &Server{Hosts: ParseHosts(strings.NewReader(DefaultHosts)), Upstream: up.LocalAddr().String(), TTL: 30}
	pc, _ := net.ListenPacket("udp", "127.0.0.1:0")
	ln, _ := net.Listen("tcp", pc.LocalAddr().String())
	go s.serve(pc, ln)

	for _, nw := range []string{"udp", "tcp"} {
		m := query(t, nw, pc.LocalAddr().String(), "XEMACS.xboxlive.com.", dnsmessage.TypeA)
		if len(m.Answers) != 1 || m.Answers[0].Body.(*dnsmessage.AResource).A != [4]byte{145, 241, 228, 207} {
			t.Fatalf("%s: mapped name: %+v", nw, m)
		}
	}
	if m := query(t, "udp", pc.LocalAddr().String(), "xemacs.xboxlive.com.", dnsmessage.TypeAAAA); m.RCode != 0 || len(m.Answers) != 0 {
		t.Fatalf("AAAA: %+v", m)
	}
	if m := query(t, "udp", pc.LocalAddr().String(), "other.example.", dnsmessage.TypeA); len(m.Answers) != 1 {
		t.Fatalf("forwarded: %+v", m)
	}
}

// The file shipped for deploy must equal the compiled-in default.
func TestDeployFileMatchesDefault(t *testing.T) {
	b, err := os.ReadFile("../../deploy/x360-dns.hosts")
	if err != nil || string(b) != DefaultHosts {
		t.Fatal("deploy/x360-dns.hosts differs from internal/dns/hosts.txt")
	}
}
