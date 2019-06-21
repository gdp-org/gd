/**
 * Copyright 2019 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package tcplib

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
		return ctx.Handler(ctx.Req)
	}

	return f.Filters[0].handle(ctx)
}
