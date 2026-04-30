#!/bin/bash

export HTTP_PROXY=http://172.21.22.151:9090 
export http_proxy=http://172.21.22.151:9090 

echo hi > hi.txt
echo you > you.txt
echo what > what.txt

oras push --plain-http --insecure 172.21.22.152:8080/oci-local/hi:v1 hi.txt
oras push --plain-http --insecure 172.21.22.152:8080/oci-local/you:v1 you.txt
oras push --plain-http --insecure 172.21.22.152:8080/oci-local/what:v1 what.txt


rm *.txt

oras pull --plain-http --insecure 172.21.22.152:8080/oci-local/hi:v1 
oras pull --plain-http --insecure 172.21.22.152:8080/oci-local/you:v1 
oras pull --plain-http --insecure 172.21.22.152:8080/oci-local/what:v1


