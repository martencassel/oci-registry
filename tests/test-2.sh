#!/bin/bash

curl -v localhost:8080/v2/
curl -I localhost:8080/v2/oci-local/alpine/blobs/sha256:989e799e634906e94dc9a5ee2ee26fc92ad260522990f26e707861a5f52bf64e
curl -I localhost:8080/v2/oci-local/alpine/blobs/sha256:a40c03cbb81c59bfb0e0887ab0b1859727075da7b9cc576a1cec2c771f38c5fb
curl -X PUT  localhost:8080/v2/oci-local/alpine/manifests/latest -d "hello"
