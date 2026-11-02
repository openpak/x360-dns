#!/bin/sh
# Idempotent: INPUT accepts for 53/udp and 53/tcp ahead of the first REJECT (rules appended after it never match).
for p in udp tcp; do
	iptables -C INPUT -p $p --dport 53 -j ACCEPT 2>/dev/null && continue
	n=$(iptables -L INPUT -n --line-numbers | awk "/REJECT/{print \$1; exit}")
	iptables -I INPUT "${n:-1}" -p $p --dport 53 -j ACCEPT
done
