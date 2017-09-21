#!/bin/sh

set -e 

CN=${1}

if [ -z ${CN}]; then
    echo "Provide a CN name for the certificates"
    exit 1
fi

## CA
# Generate root certificate private key: ca.key
openssl genrsa -out ca.key 2048

# Generate a self-signed root certificate: ca.crt
openssl req -new -key ca.key -x509 -days 3650 -out ca.crt -subj /C=CN/CN="${CN}"

## Server
# Generate a server certificate private key: server.key
openssl genrsa -out server.key 2048

# Generate a server certificate request: server.csr
openssl req -new -nodes -key server.key -out server.csr -subj /C=CN/O=Server/CN="${CN}"

# Sign the server certificate: server.crt
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt

## Client
# Generate a client certificate private key: client.key
openssl genrsa -out client.key 2048

# Generate a client certificate request: client.csr
openssl req -new -nodes -key client.key -out client.csr -subj /C=CN/O=Client/CN="${CN}"

# Sign the client certificate: client.crt
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt