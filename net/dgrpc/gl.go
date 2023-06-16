/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package dgrpc

import (
	"context"
	"github.com/chuck1024/gd/runtime/gl"
	"google.golang.org/grpc"
)

func UnaryClientCtxInterceptor() func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !gl.Exist() {
			gl.Init()
			defer gl.Close()
		}
		err := invoker(ctx, method, req, reply, cc, opts...)
		return err
	}
}

func StreamClientCtxInterceptor() func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if !gl.Exist() {
			gl.Init()
			defer gl.Close()
		}
		cs, err := streamer(ctx, desc, cc, method, opts...)
		return cs, err
	}
}

func StreamServerCtxInterceptor() func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		gl.Init()
		defer gl.Close()
		return handler(srv, ss)
	}
}

func UnaryServerCtxInterceptor() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		gl.Init()
		defer gl.Close()
		return handler(ctx, req)
	}
}

func WithGlInterceptor() InterceptorOption {
	return func(h *OptionHolder) {
		h.UnaryClientInterceptors = append(h.UnaryClientInterceptors, UnaryClientCtxInterceptor())
		h.UnaryServerInterceptors = append(h.UnaryServerInterceptors, UnaryServerCtxInterceptor())
		h.StreamClientInterceptors = append(h.StreamClientInterceptors, StreamClientCtxInterceptor())
		h.StreamServerInterceptors = append(h.StreamServerInterceptors, StreamServerCtxInterceptor())
	}
}
