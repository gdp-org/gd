/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package dogrpc

import (
	"github.com/chuck1024/gl"
	"strconv"
	"time"
)

// example: gl filter
type GlFilter struct {
	next Filter

	IfLogAll          bool
	SlowCostThreshold int
}

func (f *GlFilter) SetNext(filter Filter) {
	f.next = filter
}

func (f *GlFilter) Handle(ctx *Context) (code uint32, rsp []byte) {
	gl.Init()
	defer gl.Close()
	st := time.Now()
	logId := strconv.FormatInt(st.UnixNano(), 10)
	gl.Set(gl.LogId, logId)
	gl.Set(gl.ClientIp, ctx.ClientAddr)

	if f.next == nil {
		code, rsp = handlerWithRecover(ctx.Handler, ctx.Req)
	} else {
		code, rsp = f.next.Handle(ctx)
	}

	return code, rsp
}
