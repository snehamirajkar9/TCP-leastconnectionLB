#!/bin/sh
mkdir -p ca server client
# Generate CA cert
echo "------ Generating CA cert ------"
echo "-- Specify password for CA key"
openssl genrsa -aes256 -out ca/ca.key 4096
openssl req -new -subj "/C=CA/O=lb/CN=ca.lb.com" -x509 -sha256 -days 20 -key ca/ca.key -out ca/ca.crt

# Generate CSR for server
echo "------ Generating SERVER cert ------"
echo "-- Specify password for Server key"
openssl genrsa -aes256 -out server/server.key 4096
openssl rsa -in server/server.key -out server/server.key.pem
openssl req -new -subj "/C=CA/O=lb/CN=server.lb.com" -key server/server.key -sha256 -out server/server.csr
# Validate CSR
openssl x509 -req -days 20 -sha256 -in server/server.csr -CA ca/ca.crt -CAkey ca/ca.key -set_serial 1 -out server/server.crt

# Create client CRT
echo "------ Generating CLIENT 1 cert ------"
echo "-- Specify password for Client 1 key"
openssl genrsa -aes256 -out client/client1.key 4096
openssl rsa -in client/client1.key -out client/client1.key.pem
openssl req -new -subj "/C=CA/O=lb/CN=client1.lb.com" -key client/client1.key -out client/client1.csr
# Validate CSR
openssl x509 -req -days 20 -sha256 -in client/client1.csr -CA ca/ca.crt -CAkey ca/ca.key -set_serial 2 -out client/client1.crt

# Create client CRT
echo "------ Generating CLIENT 2 cert ------"
echo "-- Specify password for Client 2 key"
openssl genrsa -aes256 -out client/client2.key 4096
openssl rsa -in client/client2.key -out client/client2.key.pem
openssl req -new -subj "/C=CA/O=lb/CN=client2.lb.com" -key client/client2.key -out client/client2.csr
# Validate CSR
openssl x509 -req -days 20 -sha256 -in client/client2.csr -CA ca/ca.crt -CAkey ca/ca.key -set_serial 2 -out client/client2.crt
