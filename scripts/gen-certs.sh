#!/bin/bash

if [ -d "/tmp/eno-certs" ] 
then
    rm -rf /tmp/eno-certs 
fi

mkdir -p /tmp/eno-certs

pushd /tmp/eno-certs


openssl req -nodes -x509 -newkey rsa:2048 -keyout ca.key -out ca.crt -days 10000 -subj "/C=EN/ST=Eno/L=Eno/O=Eno/CN=eno" &>/dev/null
openssl req -nodes -newkey rsa:2048 -keyout server.key -out server.csr -subj "/C=EN/ST=Eno/L=Eno/O=Eno/CN=eno" &>/dev/null


openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 10000 -sha256 -extfile <(
cat <<-EOF
authorityKeyIdentifier = keyid,issuer
basicConstraints = CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = eno-webhook-service.eno-system
DNS.2 = eno-webhook-service.eno-system.svc
EOF
) &>/dev/null

CA=$(cat ca.crt | base64 -w 0)
SERVERKEY=$(cat server.key | base64 -w 0)
SERVERCRT=$(cat server.crt | base64 -w 0)
popd
sed  -i -e "s|CA_BUNDLE|$CA|g" -e "s|TLS_KEY|$SERVERKEY|g" -e "s|TLS_CRT|$SERVERCRT|" resources.yaml

rm -rf /tmp/eno-certs
