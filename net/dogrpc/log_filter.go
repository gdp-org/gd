/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package dogrpc

import (
	"encoding/json"
	"github.com/chuck1024/dlog"
	de "github.com/chuck1024/gd/derror"
	"time"
)

// example: log filter
type LogFilter struct {
	next Filter

	IfLogAll          bool
	SlowCostThreshold int
}

func (f *LogFilter) SetNext(filter Filter) {
	f.next = filter
}

func (f *LogFilter) Handle(ctx *Context) (code uint32, rsp []byte) {
	st := time.Now()

	if f.next == nil {
		code, rsp = handlerWithRecover(ctx.Handler, ctx.Req)
	} else {
		code, rsp = f.next.Handle(ctx)
	}

	cost := time.Now().Sub(st)

	logData := make(map[string]interface{})
	logData["code"] = code
	logData["ret"] = string(rsp)
	logData["cost"] = cost / time.Millisecond
	logData["args"] = string(ctx.Req)
	logData["seq"] = ctx.Seq

	logDataStr, jsonErr := json.Marshal(logData)
	if jsonErr != nil {
		dlog.Warn("logData json marshal fail, error:%s", jsonErr)
		return uint32(de.SystemError), rsp
	}

	if code != uint32(de.RpcSuccess) {
		dlog.WarnT("SESSION", "%s %s", ctx.Method, logDataStr)
	} else {
		dlog.InfoT("SESSION", "%s %s", ctx.Method, logDataStr)
	}

	if f.SlowCostThreshold > 0 && cost > time.Duration(f.SlowCostThreshold)*time.Millisecond {
		dlog.WarnT("SERVER_SLOW", "%s %s", ctx.Method, logDataStr)
	}

	return code, rsp
}
