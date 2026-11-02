# Protocol: x360-dns

Plain DNS (RFC 1035) over UDP and TCP on port 53, A records only.

## Names

The map is a hosts-style file: `<ipv4> <name> [<name>...]`, `#` comments. Names are matched case-insensitively.
`deploy/x360-dns.hosts` is a copy of the compiled-in `internal/dns/hosts.txt` (a test keeps them equal).

Secondary (145.241.228.207: kdc, gateway, xexds): `xemacs`, `xeas`, `xetgs`, `xexds` `.xboxlive.com`.

Primary (145.241.199.19, Traefik), served today: `xncsi`, `progress`, `profile`, `content`, `xuacs`,
`xnotify-production-origin` `.xboxlive.com`; `notice.xbox.com`; `nexus.passport.com`; `login.live.com`.

Primary, NOT served yet (found in XAM 17559 strings or the handoff notes): `sts.xboxlive.com`, `xam.xboxlive.com`,
`xlink.xboxlive.com`, `msac.xboxlive.com`, `macs.xbox.com`. They resolve to the primary so a console fails against
OpenPak rather than reaching Microsoft.

## Behaviour

- Mapped name, `A`: one answer, TTL 30. `AAAA`: NOERROR, empty. Any other type for a mapped name is forwarded.
- Unmapped name: the query is forwarded unchanged to the upstream resolver; its reply is returned. Upstream failure
  gives SERVFAIL.
- TCP is length-prefixed; a truncated UDP reply from upstream makes the client retry over TCP, which is forwarded over TCP.
