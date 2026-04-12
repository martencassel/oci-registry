#!/bin/bash
set -euo pipefail

IP_ADDR="172.21.22.152"
PORT="9090"
NIC="eth0"

ensure_ip() {
    ip addr show "$NIC" | grep -q "$IP_ADDR" \
        || sudo -E ipmgr add "$IP_ADDR" --iface "$NIC"
}

ensure_ip
go run .
