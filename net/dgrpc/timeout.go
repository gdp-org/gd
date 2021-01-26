/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package dgrpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

func WithClientTimeOutInterceptor(defaultTimeout time.Duration) InterceptorOption {
	return func(h *OptionHolder) {
		h.UnaryClientInterceptors = append(h.UnaryClientInterceptors, UnaryClientTimeOutInterceptor(defaultTimeout))
		h.StreamClientInterceptors = append(h.StreamClientInterceptors, StreamClientTimeOutInterceptor(defaultTimeout))
	}
}

func UnaryClientTimeOutInterceptor(defaultTimeout time.Duration) func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
			defer cancel()
		}
		err := invoker(ctx, method, req, reply, cc, opts...)
		return err
	}
}

func StreamClientTimeOutInterceptor(defaultTimeout time.Duration) func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
			defer cancel()
		}
		cs, err := streamer(ctx, desc, cc, method, opts...)
		return cs, err
	}
}
