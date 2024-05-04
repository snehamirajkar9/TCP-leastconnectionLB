#!/bin/sh
CERTS_FOLDER=$1

SERVER_IP=127.0.0.1
SERVER_NAME=server.lb.com
PORT=9443
DST=$SERVER_NAME:$PORT
CLIENT_CERT=${CERTS_FOLDER}/client/client1.crt
CLIENT_KEY=${CERTS_FOLDER}/client/client1.key.pem
CLIENT_CERT_ALT=${CERTS_FOLDER}/client/client2.crt
CLIENT_KEY_ALT=${CERTS_FOLDER}/client/client2.key.pem
CA_CERT=${CERTS_FOLDER}/ca/ca.crt
CERT_PARAMS="--cert ${CLIENT_CERT} --key ${CLIENT_KEY} --cacert ${CA_CERT}"

CERT_PARAMS_ALT="--cert ${CLIENT_CERT_ALT} --key ${CLIENT_KEY_ALT} --cacert ${CA_CERT}"

curl --resolve ${DST}:${SERVER_IP} --tlsv1.3 ${CERT_PARAMS} https://${DST}/ 
curl --resolve ${DST}:${SERVER_IP} --tlsv1.3 ${CERT_PARAMS_ALT} https://${DST}/