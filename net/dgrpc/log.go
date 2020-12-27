/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package dgrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/log4go"
	log "github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/runtime/gl"
	"github.com/chuck1024/gd/utls"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

/*
	GRPC LOG Level copy from package:grpc log
*/
const (
	// infoLog indicates Info severity.
	infoLog int = iota
	// warningLog indicates Warning severity.
	warningLog
	// errorLog indicates Error severity.
	errorLog
	// fatalLog indicates Fatal severity.
	fatalLog
)

func WithLogInterceptor() InterceptorOption {
	return func(h *OptionHolder) {
		h.UnaryServerInterceptors = append(h.UnaryServerInterceptors, UnaryServerLoggerInterceptor())
		h.StreamServerInterceptors = append(h.StreamServerInterceptors, StreamServerLoggerInterceptor())
	}
}

func StreamServerLoggerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		st := time.Now()

		ip := GetClientIP(stream.Context())
		if ip != "" {
			gl.Set(gl.ClientIp, ip)
		}

		gl.Set(gl.Url, info.FullMethod)
		logData := make(map[string]interface{})

		err := handler(srv, stream)
		cost := time.Now().Sub(st)
		code := status.Code(err)
		logData["code"] = code.String()
		if err != nil {
			logData["err"] = err.Error()
		}

		costMs := cost / time.Millisecond
		logData["cost"] = costMs
		if costMs >= 50 || err != nil {
			logData["ctx"] = gl.JsonCurrentCtx()
		}

		logDataStr, jsonErr := json.Marshal(logData)
		if jsonErr != nil {
			log.Warn("logData json marshal fail, error:%s", jsonErr)
			return err
		}

		if ShouldFail4Code(code) {
			log.WarnT("SESSION", fmt.Sprintf("%s %s", info.FullMethod, logDataStr))
		} else {
			log.WarnT("SESSION", fmt.Sprintf("%s %s", info.FullMethod, logDataStr))
		}
		return err
	}
}

func UnaryServerLoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(context context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		st := time.Now()
		ip := GetClientIP(context)
		if ip != "" {
			gl.Set(gl.ClientIp, ip)
		}
		gl.Set(gl.Url, info.FullMethod)

		logData := make(map[string]interface{})
		logData["args"] = req

		resp, err := handler(context, req)

		cost := time.Now().Sub(st)
		costMs := cost / time.Millisecond
		if costMs > 50 || err != nil {
			ctxJson := gl.JsonCurrentCtx()
			logData["ctx"] = ctxJson
		}

		logData["ret"] = resp
		code := status.Code(err)
		logData["code"] = code.String()
		if err != nil {
			logData["err"] = err.Error()
		}

		logData["cost"] = costMs
		logDataStr, jsonErr := json.Marshal(logData)
		if jsonErr != nil {
			log.Warn("logData json marshal fail, error:%s", jsonErr)
			return resp, err
		}
		if ShouldFail4Code(code) {
			log.WarnT("SESSION", fmt.Sprintf("%s %s", info.FullMethod, logDataStr))
		} else {
			log.InfoT("SESSION", fmt.Sprintf("%s %s", info.FullMethod, logDataStr))
		}
		return resp, err
	}
}

/*
	GetGrpcLogger,use same logger with dlog
*/

var Logger = &GrpcLogger{}

const LogTag = "GRPC_PROCESS"

type GrpcLogger struct{}

func SetGrpcLogger() {
	grpclog.SetLoggerV2(Logger)
}

func (l *GrpcLogger) Info(args ...interface{}) {
	logStr := fmt.Sprint(args...)
	l.LogWithDepthAndLevel(log4go.INFO, logStr)
}

func (l *GrpcLogger) Infoln(args ...interface{}) {
	logStr := fmt.Sprint(args...)
	l.LogWithDepthAndLevel(log4go.INFO, logStr)
}

func (l *GrpcLogger) Infof(format string, args ...interface{}) {
	logStr := fmt.Sprintf(format, args...)
	l.LogWithDepthAndLevel(log4go.INFO, logStr)
}

func (l *GrpcLogger) Warning(args ...interface{}) {
	logStr := fmt.Sprint(args...)
	l.LogWithDepthAndLevel(log4go.WARNING, logStr)
}

func (l *GrpcLogger) Warningln(args ...interface{}) {
	logStr := fmt.Sprint(args...)
	l.LogWithDepthAndLevel(log4go.WARNING, logStr)
}

func (l *GrpcLogger) Warningf(format string, args ...interface{}) {
	logStr := fmt.Sprintf(format, args...)
	l.LogWithDepthAndLevel(log4go.WARNING, logStr)
}

func (l *GrpcLogger) Error(args ...interface{}) {
	logStr := fmt.Sprint(args...)
	l.LogWithDepthAndLevel(log4go.ERROR, logStr)
}

func (l *GrpcLogger) Errorln(args ...interface{}) {
	logStr := fmt.Sprint(args...)
	l.LogWithDepthAndLevel(log4go.ERROR, logStr)
}

func (l *GrpcLogger) Errorf(format string, args ...interface{}) {
	logStr := fmt.Sprintf(format, args...)
	l.LogWithDepthAndLevel(log4go.ERROR, logStr)
}

func (l *GrpcLogger) Fatal(args ...interface{}) {
	log.Crash(args...)
}

func (l *GrpcLogger) Fatalln(args ...interface{}) {
	log.Crash(args...)
}

func (l *GrpcLogger) Fatalf(format string, args ...interface{}) {
	log.Crash(args...)
}

func (l *GrpcLogger) V(level int) bool {
	switch level {
	case infoLog:
		return log.IsEnabledFor(log.INFO)
	case warningLog:
		return log.IsEnabledFor(log.WARNING)
	case errorLog:
		return log.IsEnabledFor(log.ERROR)
	case fatalLog:
		return true
	default:
		return false
	}
}

func (l *GrpcLogger) LogWithDepthAndLevel(level log4go.Level, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	url, ip, logId := batchGetCtx()
	first := utls.MustString(args[0], "")
	if len(args) > 1 {
		log.Global.IntLogfTagUrl(LogTag, ip, logId, url, log.Level(level), first, args[1:]...)
	} else {
		log.Global.IntLogfTagUrl(LogTag, ip, logId, url, log.Level(level), first)
	}
}

func batchGetCtx() (url, ip, logId string) {
	vs := gl.BatchGet([]interface{}{
		gl.LogId,
		gl.Url,
		gl.ClientIp,
	})
	if vs == nil {
		return
	}

	urlo, ok := vs[gl.Url]
	if ok && urlo != nil {
		url, _ = urlo.(string)
	}

	ipo, ok := vs[gl.ClientIp]
	if ok && ipo != nil {
		ip, _ = ipo.(string)
	}

	logIdo, ok := vs[gl.LogId]
	if ok && logIdo != nil {
		logId, _ = logIdo.(string)
	}
	return
}
