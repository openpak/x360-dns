# Deploy

Secondary host (145.241.228.207, x86_64), quadlet `deploy/openpak-x360-dns.container`, `Network=host`, listening on
the box's private address 10.0.0.22:53 (systemd-resolved owns only the 127.0.0.x stubs; the primary cannot host it,
nn-sssl-dns holds 53 there).

One-time on the secondary (as root): `deploy/99-openpak-dns.conf` to `/etc/sysctl.d/` (`sysctl --system`), `deploy/x360-dns-fw.sh`
and `deploy/openpak-x360-dns-fw.service` to `/etc/openpak/` and `/etc/systemd/system/` (enable --now). Then
`deploy/x360-dns.hosts` to `~/openpak/deploy/x360-dns.hosts`, the quadlet to `~/.config/containers/systemd/`, and
`systemctl --user start openpak-x360-dns`. Edit the hosts file and restart the unit to change the map without a rebuild.

Verify from outside: `dig @145.241.228.207 xemacs.xboxlive.com` (145.241.228.207) and `dig @145.241.228.207 example.com`.
Consoles: Settings, Network, Configure Network, DNS Settings, Manual, primary DNS 145.241.228.207.

Live since v0.1.0. 53/udp is reachable from the internet; 53/tcp answers on the box but the cloud security list drops it from outside (consoles use UDP; open TCP 53 in the OCI security list for the secondary to enable truncation fallback).
