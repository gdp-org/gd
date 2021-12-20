/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package dogrpc

import (
	"fmt"
	"github.com/chuck1024/gd/dlog"
	"runtime"
)

var globalFilter = &Filters{}

type Filter interface {
	SetNext(filter Filter)
	Handle(cxt *Context) (uint32, []byte)
}

type Filters struct {
	Filters []Filter
}

func Use(filters []Filter) {
	var front Filter
	for _, filter := range filters {
		globalFilter.Filters = append(globalFilter.Filters, filter)
		if front != nil {
			front.SetNext(filter)
		}
		front = filter
	}
}

func (f *Filters) Handle(ctx *Context) (uint32, []byte) {
	if len(f.Filters) == 0 {
		return handlerWithRecover(ctx.Handler, ctx.Req)
	}

	return f.Filters[0].Handle(ctx)
}

func handlerWithRecover(f RpcHandlerFunc, req []byte) (code uint32, resp []byte) {
	defer func() {
		if x := recover(); x != nil {
			code = uint32(InternalServerError.Code())
			stackTrace := make([]byte, 1<<20)
			n := runtime.Stack(stackTrace, false)
			errStr := fmt.Sprintf("Panic occured: %v\n Stack trace: %s", x, stackTrace[:n])
			dlog.Error("handlerWithRecover occur error:%s", errStr)
		}
	}()

	code, resp = f(req)
	return
}
