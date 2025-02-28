# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
rm -rf secrets
mkdir secrets
echo subjectAltName = DNS:localhost,DNS:docker-local-sdncontroller-1,DNS:docker-local-psql-1,DNS:docker-local-sdncontroller-1,IP:127.0.0.1 > secrets/extfile.cnf

## Set up CA
openssl genrsa -out secrets/ca.key 2048
openssl req -new -subj "/C=US/ST=CA/O=IDC.dev/CN=SDN-Controller-Global.dev" -key secrets/ca.key -out secrets/ca.csr
openssl x509 -req -days 365 -in secrets/ca.csr -signkey secrets/ca.key -out secrets/cacert.pem

## Secrets for SQL
openssl genrsa -out secrets/sql-server-privkey.pem 2048
openssl req -new -subj "/C=US/ST=CA/O=IDC.dev/CN=SQL-Server" -key secrets/sql-server-privkey.pem -out secrets/sql-server.csr
openssl x509 -req -in secrets/sql-server.csr -CA secrets/cacert.pem -CAkey secrets/ca.key -CAcreateserial -out secrets/sql-server-cert.pem -days 500 -sha256 -extfile secrets/extfile.cnf
echo "foo" > secrets/sql-passwd

## Secrets for OVN
openssl genrsa -out secrets/ovn-central-privkey.pem 2048
openssl req -new -subj "/C=US/ST=CA/O=IDC.dev/CN=OVN-Central" -key secrets/ovn-central-privkey.pem -out secrets/ovn-central.csr
openssl x509 -req -in secrets/ovn-central.csr -CA secrets/cacert.pem -CAkey secrets/ca.key -CAcreateserial -out secrets/ovn-central-cert.pem -days 500 -sha256 -extfile secrets/extfile.cnf

## Secrets for SDN Controller
openssl genrsa -out secrets/sdn-controller-privkey.pem 2048
openssl req -new -subj "/C=US/ST=CA/O=IDC.dev/CN=SDN-Controller" -key secrets/sdn-controller-privkey.pem -out secrets/sdn-controller.csr
openssl x509 -req -in secrets/sdn-controller.csr -CA secrets/cacert.pem -CAkey secrets/ca.key -CAcreateserial -out secrets/sdn-controller-cert.pem -days 500 -sha256 -extfile secrets/extfile.cnf
