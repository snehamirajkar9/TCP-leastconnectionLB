#!/bin/sh

# Second CA to test rejection
openssl genrsa -aes256 -out ca/ca-alt.key 4096
openssl req -new -subj "/C=CA/O=lb/CN=ca.lb.com" -x509 -sha256 -days 20 -key ca/ca-alt.key -out ca/ca-alt.crt
# Other client with alternative CA
openssl genrsa -aes256 -out client/client-alt.key 4096
openssl rsa -in client/client-alt.key -out client/client-alt.key.pem
openssl req -new -subj "/C=CA/O=lb/CN=clientalt.lb.com" -key client/client-alt.key -out client/client-alt.csr
# Validate CSR
openssl x509 -req -days 20 -sha256 -in client/client-alt.csr -CA ca/ca-alt.crt -CAkey ca/ca-alt.key -set_serial 2 -out client/client-alt.crt

