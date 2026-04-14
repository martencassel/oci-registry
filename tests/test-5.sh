#!/bin/bash


podman image pull docker.io/postgres:latest
podman image tag docker.io/postgres:latest  172.21.22.152:8080/oci-local/postgres:latest
HTTP_PROXY=http://172.21.22.151:9090 http_proxy=http://172.21.22.151:9090 podman image push \
	172.21.22.152:8080/oci-local/postgres:latest --tls-verify=false  

