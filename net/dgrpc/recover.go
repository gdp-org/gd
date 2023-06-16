/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package dgrpc

import (
	"fmt"
	"github.com/chuck1024/gd/dlog"
	"github.com/go-errors/errors"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
)

func WithRecoveryInterceptor(handler func(p interface{}) (err error)) InterceptorOption {
	if handler == nil {
		handler = DefaultRecoveryHandler
	}

	option := grpc_recovery.WithRecoveryHandler(handler)
	return func(h *OptionHolder) {
		h.UnaryServerInterceptors = append(h.UnaryServerInterceptors, grpc_recovery.UnaryServerInterceptor(option))
		h.StreamServerInterceptors = append(h.StreamServerInterceptors, grpc_recovery.StreamServerInterceptor(option))
	}
}

func DefaultRecoveryHandler(p interface{}) (err error) {
	wrap := errors.Wrap(p, 2)
	stacktrace := wrap.ErrorStack()
	dlog.Critical("panic_recoverd!" + stacktrace)
	fmt.Fprintln(os.Stderr, "panic_recovered:", stacktrace)

	return status.Errorf(codes.Internal, "%s", p)
}
