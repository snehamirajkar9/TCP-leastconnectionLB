#!/bin/sh
CERTS_FOLDER=$1

SERVER_IP=127.0.0.1
SERVER_NAME=server.lb.com
PORT=9443
DST=$SERVER_NAME:$PORT
CLIENT_CERT_ALT=${CERTS_FOLDER}/client/client-alt.crt
CLIENT_KEY_ALT=${CERTS_FOLDER}/client/client-alt.key.pem
CA_CERT_ALT=${CERTS_FOLDER}/ca/ca-alt.crt

CERT_PARAMS_ALT="--cert ${CLIENT_CERT_ALT} --key ${CLIENT_KEY_ALT} --cacert ${CA_CERT_ALT}"

curl --resolve ${DST}:${SERVER_IP} --tlsv1.3 ${CERT_PARAMS_ALT} https://${DST}/ 