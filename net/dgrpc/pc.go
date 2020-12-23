/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package dgrpc

import (
	"context"
	"fmt"
	"github.com/chuck1024/gd/runtime/gl"
	"github.com/chuck1024/gd/runtime/pc"
	"github.com/chuck1024/gd/utls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
)

var (
	metaTraceId = "_traceId"
)

//to ensure traceid in all context
//a simple tracing like opentracing
func ensureTraceId(context context.Context) context.Context {
	traceId := ""
	md, ok := metadata.FromIncomingContext(context)
	if ok && md != nil {
		traceIdMeta := md[metaTraceId]
		if len(traceIdMeta) > 0 {
			traceId = traceIdMeta[0]
		}
	}
	if md == nil {
		md = metadata.Pairs()
	}
	if traceId == "" {
		logId, _ := gl.Get(gl.LogId)
		traceId = utls.MustString(logId, utls.TraceId())
		md.Set(metaTraceId, traceId)
		context = metadata.NewOutgoingContext(context, md)
	}
	gl.Set(metaTraceId, traceId)
	gl.Set(gl.LogId, traceId)
	return context
}

func ensureTraceIdStream(ss grpc.ServerStream) {
	traceId := ""
	ctx := ss.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if ok && md != nil {
		traceIdMeta := md[metaTraceId]
		if len(traceIdMeta) > 0 {
			traceId = traceIdMeta[0]
		}
	}
	if traceId == "" {
		logId, _ := gl.Get(gl.LogId)
		traceId = utls.MustString(logId, utls.TraceId())
		ss.SetHeader(metadata.Pairs(metaTraceId, traceId))
	}
	gl.Set(metaTraceId, traceId)
	gl.Set(gl.LogId, traceId)
}

func UnaryClientPerfCounterInterceptor(service string) func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		st := time.Now()
		ctx = ensureTraceId(ctx)
		err := invoker(ctx, method, req, reply, cc, opts...)
		cost := time.Now().Sub(st)
		k := fmt.Sprintf("service=%v,method=%v,st=client", service, method)
		pc.Cost(k, cost)
		ctxKey := method
		gl.IncrCountKey(ctxKey, 1)
		gl.IncrCostKey(ctxKey, cost)
		if err != nil {
			pc.CostFail(k, 1)
			gl.IncrFailKey(ctxKey, 1)
		}
		return err
	}
}

func StreamClientPerfCounterInterceptor(service string) func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		st := time.Now()
		ctx = ensureTraceId(ctx)
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		cost := time.Now().Sub(st)
		k := fmt.Sprintf("service=%v,method=%v,st=client", service, method)
		pc.Cost(k, cost)
		ctxKey := method
		gl.IncrCountKey(ctxKey, 1)
		gl.IncrCostKey(ctxKey, cost)
		if err != nil {
			pc.CostFail(k, 1)
			gl.IncrFailKey(ctxKey, 1)
		}
		return clientStream, err
	}
}

func StreamServerPerfCounterInterceptor(service string) func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		st := time.Now()
		ensureTraceIdStream(ss)
		m := info.FullMethod
		err := handler(srv, ss)
		cost := time.Now().Sub(st)
		k := fmt.Sprintf("service=%v,method=%v,st=server", service, m)
		pc.Cost(k, cost)
		if err != nil {
			pc.CostFail(k, 1)
		}
		return err
	}
}

func UnaryServerPerfCounterInterceptor(service string) func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		st := time.Now()
		ctx = ensureTraceId(ctx)
		m := info.FullMethod
		resp, err := handler(ctx, req)
		cost := time.Now().Sub(st)
		k := fmt.Sprintf("service=%v,method=%v,st=server", service, m)
		pc.Cost(k, cost)
		if err != nil {
			pc.CostFail(k, 1)
		}
		return resp, err
	}
}

func WithPerfCounterInterceptor(service string) InterceptorOption {
	return func(h *OptionHolder) {
		h.UnaryClientInterceptors = append(h.UnaryClientInterceptors, UnaryClientPerfCounterInterceptor(service))
		h.UnaryServerInterceptors = append(h.UnaryServerInterceptors, UnaryServerPerfCounterInterceptor(service))
		h.StreamClientInterceptors = append(h.StreamClientInterceptors, StreamClientPerfCounterInterceptor(service))
		h.StreamServerInterceptors = append(h.StreamServerInterceptors, StreamServerPerfCounterInterceptor(service))
	}
}
