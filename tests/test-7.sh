#!/bin/bash


export HTTP_PROXY=http://172.21.22.151:9090
export http_proxy=http://172.21.22.151:9090

echo hi > hi.txt

podman image tag alpine:latest  172.21.22.152:8080/oci-local/hello:v1
podman image push 172.21.22.152:8080/oci-local/hello:v1 --tls-verify=false  --log-level=trace

oras attach --plain-http --insecure --artifact-type doc/example 172.21.22.152:8080/oci-local/hello:v1 hi.txt

