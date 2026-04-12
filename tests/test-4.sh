#!/bin/bash

HTTP_PROXY=http://172.21.22.151:9090 http_proxy=http://172.21.22.151:9090 oras push 172.21.22.152:8080/oci-local/hello:v1 hello.txt  --plain-http --debug
