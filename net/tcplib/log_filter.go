/**
 * Copyright 2019 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package tcplib

import (
	"encoding/json"
	"github.com/chuck1024/doglog"
	de "github.com/chuck1024/godog/error"
	"time"
)

// example: log filter
type LogFilter struct {
	next Filter

	IfLogAll          bool
	SlowCostThreshold int
}

func (f *LogFilter) setNext(filter Filter) {
	f.next = filter
}

func (f *LogFilter) handle(ctx *Context) (code uint32, rsp []byte) {
	st := time.Now()

	if f.next == nil {
		code, rsp = ctx.Handler(ctx.Req)
	} else {
		code, rsp = f.next.handle(ctx)
	}

	cost := time.Now().Sub(st)

	logData := make(map[string]interface{})
	logData["code"] = code
	logData["ret"] = string(rsp)
	logData["cost"] = cost / time.Millisecond

	logDataStr, jsonErr := json.Marshal(logData)
	if jsonErr != nil {
		doglog.Warn("logData json marshal fail, error:%s", jsonErr)
		return uint32(de.SystemError), rsp
	}

	if code != uint32(de.TcpSuccess) {
		doglog.Warn("SESSION %s", logDataStr)
	} else {
		doglog.Info("SESSION %s", logDataStr)
	}

	if f.SlowCostThreshold > 0 && cost > time.Duration(f.SlowCostThreshold)*time.Millisecond {
		doglog.Warn("SERVER_SLOW %s", logDataStr)
	}

	return code, rsp
}
