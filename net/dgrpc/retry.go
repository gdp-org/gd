/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package dgrpc

import "github.com/grpc-ecosystem/go-grpc-middleware/retry"

func WithRetryInterceptor(option ...grpc_retry.CallOption) InterceptorOption {
	return func(h *OptionHolder) {
		h.UnaryClientInterceptors = append(h.UnaryClientInterceptors, grpc_retry.UnaryClientInterceptor(option...))
		h.StreamClientInterceptors = append(h.StreamClientInterceptors, grpc_retry.StreamClientInterceptor(option...))
	}
}