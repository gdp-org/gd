#!/bin/bash

servicename=helloworld

# to install grpc-go and grpc-gateway
#go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
#go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
#go get -u github.com/golang/protobuf/protoc-gen-go
#go get -u github.com/go-bindata/go-bindata/...

# the pb go file
protoc -I/usr/local/include -I. \
  --go_out=plugins=grpc:. \
  ${servicename}.proto

# the gateway go file
# protoc -I/usr/local/include -I. \
  #--grpc-gateway_out=logtostderr=true,grpc_api_configuration=${servicename}.yaml:. \
  #${servicename}.proto
