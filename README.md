# Xbox 360 Console DNS

OpenPak's DNS server for the Xbox 360. A console has no hosts file; its only knob is the DNS server in the network
settings. This server answers the Xbox Live hostnames with the right OpenPak box and forwards everything else.

| | |
| --- | --- |
| Network | Xbox 360 |
| Role | Console DNS (53/udp + 53/tcp) |
| Depends on | `golang.org/x/net/dns/dnsmessage` (vendored) |
| Image | `ghcr.io/openpak/x360-dns` |
| Used by | Xbox 360 consoles (set the DNS server to 145.241.228.207) |
| Licence | AGPL-3.0-only |

## What it serves

`A` queries for the names in the hosts file (`internal/dns/hosts.txt`, compiled in; `X360_DNS_HOSTS` points at an
override read at start). `AAAA` for those names is NOERROR with no data. Everything else goes to `X360_DNS_UPSTREAM`
(default `1.1.1.1:53`) over the transport the client used. Answers carry a 30 s TTL; each answered name is logged
at most once a minute.

## Run

```sh
GOWORK=off go run -mod=vendor ./cmd/x360-dns     # X360_DNS_ADDR defaults to :53
```

## Docs

- [Protocol](docs/protocol.md)
- [Deploy](docs/deploy.md)

## Licence

AGPL-3.0-only. See [LICENSE](LICENSE).
