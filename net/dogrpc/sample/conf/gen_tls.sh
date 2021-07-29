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
openssl req -new -x509 -key server.key -out server.pem -days 3650

# 客户端证书
openssl genrsa -out client.key 2048
openssl req -new -x509 -key client.key -out client.pem -days 3650
