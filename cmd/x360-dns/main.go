// x360-dns answers Xbox Live hostnames with OpenPak's boxes and forwards everything else.
package main

import (
	"log"
	"os"

	"github.com/openpak/x360-dns/internal/dns"
)

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	hosts, err := dns.LoadHosts(os.Getenv("X360_DNS_HOSTS"))
	if err != nil {
		log.Fatal(err)
	}
	s := &dns.Server{Hosts: hosts, Upstream: env("X360_DNS_UPSTREAM", "1.1.1.1:53"), TTL: 30}
	addr := env("X360_DNS_ADDR", ":53")
	log.Printf("x360-dns on %s, %d names, upstream %s", addr, len(hosts), s.Upstream)
	log.Fatal(s.Serve(addr))
}
