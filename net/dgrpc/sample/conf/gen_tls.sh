#!/bin/bash

# 根证书
openssl genrsa -out ca.key 2048
openssl req -new -x509 -key ca.key -out ca.pem -days 36500

# Country Name (2 letter code) []:
# State or Province Name (full name) []:
# Locality Name (eg, city) []:
# Organization Name (eg, company) []:
# Organizational Unit Name (eg, section) []:
# Common Name (eg, fully qualified host name) []:gd
# Email Address []:

# 服务端证书
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr
openssl x509 -req -sha256 -CA ca.pem -CAkey ca.key -CAcreateserial -days 36500 -in server.csr -out server.pem

# 客户端证书
openssl ecparam -genkey -name secp384r1 -out client.key
openssl req -new -key client.key -out client.csr
openssl x509 -req -sha256 -CA ca.pem -CAkey ca.key -CAcreateserial -days 3650 -in client.csr -out client.pem
