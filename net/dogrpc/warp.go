/**
 * Copyright 2019 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package dogrpc

import (
	"encoding/json"
	"fmt"
	"github.com/chuck1024/doglog"
	de "github.com/chuck1024/godog/error"
	"reflect"
)

var errInterface = reflect.TypeOf((*error)(nil)).Elem()

func wrap(toWrap interface{}) (RpcHandlerFunc, error) {
	refToWrap := reflect.ValueOf(toWrap)
	wt := reflect.TypeOf(toWrap)
	if wt.Kind() != reflect.Func {
		return nil, fmt.Errorf("toWrap must be func,type=%v,func=%v", wt, toWrap)
	}
	wtNumIn := wt.NumIn()
	if wtNumIn < 1 {
		return nil, fmt.Errorf("params in count must > 1 %v", toWrap)
	}
	inType := wt.In(0)
	if wt.NumOut() < 4 {
		return nil, fmt.Errorf("params out count must > 4 %v", toWrap)
	}
	if wt.Out(0).Kind() != reflect.Uint32 {
		return nil, fmt.Errorf("params out 1 must be uint32 %v", toWrap)
	}
	if wt.Out(1).Kind() != reflect.String {
		return nil, fmt.Errorf("params out 2 must be string %v", toWrap)
	}
	if wt.Out(2).Kind() != reflect.Interface {
		return nil, fmt.Errorf("params out 3 must be interface %v", toWrap)
	}
	if !wt.Out(2).Implements(errInterface) {
		return nil, fmt.Errorf("params out 4 must be error %v", toWrap)
	}

	wrapped := func(req []byte) (code uint32, resp []byte) {
		var inVal reflect.Value
		if inType.Kind() == reflect.Ptr {
			ite := inType.Elem()
			inVal = reflect.New(ite)
		} else {
			inVal = reflect.New(inType).Elem()
		}

		in := make([]reflect.Value, wtNumIn)
		in[0] = inVal
		out := refToWrap.Call(in)
		if len(out) != 4 {
			doglog.Error("warp return not 4!in=%v,out=%v,func=%v", in, out, toWrap)
			code = uint32(de.RpcInternalServerError)
			resp = Return(uint32(de.RpcInternalServerError), "ret not 4!", nil, nil)
			return
		}

		var (
			message string
			err     error
			ret     interface{}
		)

		if out[0].CanInterface() {
			code, _ = out[0].Interface().(uint32)
		} else {
			code = uint32(de.RpcInternalServerError)
			doglog.Error("warp not parse code!in=%v,out=%v,func=%v", in, out, toWrap)
		}

		if out[1].CanInterface() {
			message, _ = out[1].Interface().(string)
		} else {
			doglog.Error("warp not parse message!in=%v,out=%v,func=%v", in, out, toWrap)
		}

		if out[2].CanInterface() {
			err, _ = out[2].Interface().(error)
		} else {
			doglog.Error("warp not parse err!in=%v,out=%v,func=%v", in, out, toWrap)
		}
		if out[3].CanInterface() {
			ret = out[3].Interface()
		} else {
			doglog.Error("warp not parse result!in=%v,out=%v,func=%v", in, out, toWrap)
		}

		doglog.Debug("warp wrapped call,in=%v,out=%v,func=%v", in, out, toWrap)
		resp = Return(code, message, err, ret)
		return
	}

	return wrapped, nil
}

func Return(code uint32, message string, err error, result interface{}) (resp []byte) {
	ret := make(map[string]interface{})
	ret["code"] = code
	ret["result"] = result
	ret["message"] = message

	resp, _ = json.Marshal(ret)
	return
}
