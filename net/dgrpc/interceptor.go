/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package dgrpc

import (
	"google.golang.org/grpc"
)

type InterceptorOption func(*OptionHolder)

func GetOptionHolder(option ...InterceptorOption) *OptionHolder {
	holder := new(OptionHolder)
	for _, v := range option {
		v(holder)
	}
	return holder
}

type OptionHolder struct {
	UnaryServerInterceptors  []grpc.UnaryServerInterceptor
	StreamServerInterceptors []grpc.StreamServerInterceptor
	UnaryClientInterceptors  []grpc.UnaryClientInterceptor
	StreamClientInterceptors []grpc.StreamClientInterceptor
}