/**
 * Copyright 2019 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package dogrpc

import (
	"fmt"
	"github.com/chuck1024/doglog"
	"runtime"
)

type Filter interface {
	setNext(filter Filter)
	handle(cxt *Context) (uint32, []byte)
}

type Filters struct {
	Filters []Filter
}

var GF = &Filters{}

func InitFilters(filters []Filter) {
	var frontFilter Filter
	for _, filter := range filters {
		GF.Filters = append(GF.Filters, filter)
		if frontFilter != nil {
			frontFilter.setNext(filter)
		}
		frontFilter = filter
	}
}

func (f *Filters) Handle(ctx *Context) (uint32, []byte) {
	if len(f.Filters) == 0 {
		return handlerWithRecover(ctx.Handler, ctx.Req)
	}

	return f.Filters[0].handle(ctx)
}

func handlerWithRecover(f RpcHandlerFunc, req []byte) (code uint32, resp []byte) {
	defer func() {
		if x := recover(); x != nil {
			code = uint32(InternalServerError.Code())
			stackTrace := make([]byte, 1<<20)
			n := runtime.Stack(stackTrace, false)
			errStr := fmt.Sprintf("Panic occured: %v\n Stack trace: %s", x, stackTrace[:n])
			doglog.Error("[handlerWithRecover] occur error:%s", errStr)
		}
	}()

	code, resp = f(req)
	return
}
