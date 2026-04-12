#!/bin/bash


podman image pull alpine:latest
podman image tag alpine:latest  172.21.22.152:8080/oci-local/alpine:latest
HTTP_PROXY=http://172.21.22.151:9090 http_proxy=http://172.21.22.151:9090 podman image push \
	172.21.22.152:8080/oci-local/alpine:latest --tls-verify=false  --log-level=trace

