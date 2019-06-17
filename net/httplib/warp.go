/**
 * Copyright 2019 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package httplib

import (
	"encoding/json"
	"fmt"
	"github.com/chuck1024/doglog"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"reflect"
)

var ptrToGinCtx = reflect.PtrTo(reflect.TypeOf((*gin.Context)(nil))).Kind()
var errInterface = reflect.TypeOf((*error)(nil)).Elem()

// example: warp to gin.HandlerFunc -- func(*Context)
func Wrap(toWrap interface{}) (gin.HandlerFunc, error) {
	refToWrap := reflect.ValueOf(toWrap)
	wt := reflect.TypeOf(toWrap)
	if wt.Kind() != reflect.Func {
		return nil, fmt.Errorf("toWrap must be func,type=%v,func=%v", wt, toWrap)
	}
	wtNumIn := wt.NumIn()
	if wtNumIn < 2 {
		return nil, fmt.Errorf("params in count must > 2 %v", toWrap)
	}
	if wt.In(0).Kind() != ptrToGinCtx {
		return nil, fmt.Errorf("first param in must *gin.Context %v", toWrap)
	}
	inType := wt.In(1)
	if wt.NumOut() < 4 {
		return nil, fmt.Errorf("params out count must > 4 %v", toWrap)
	}
	if wt.Out(0).Kind() != reflect.Int {
		return nil, fmt.Errorf("params out 1 must be int %v", toWrap)
	}
	if wt.Out(1).Kind() != reflect.String {
		return nil, fmt.Errorf("params out 2 must be string %v", toWrap)
	}
	if wt.Out(2).Kind() != reflect.Interface {
		return nil, fmt.Errorf("params out 3 must be interface %v", toWrap)
	}
	if !wt.Out(2).Implements(errInterface) {
		return nil, fmt.Errorf("params out 3 must be error %v", toWrap)
	}

	wrapped := func(c *gin.Context) {
		var inVal reflect.Value
		if inType.Kind() == reflect.Ptr {
			ite := inType.Elem()
			inVal = reflect.New(ite)
		} else {
			inVal = reflect.New(inType).Elem()
		}
		inValInterface := inVal.Interface()

		// parse data
		// data_raw is possible to encrypt data
		dataBtsObj, ok := c.Get(DATA_RAW)
		if !ok {
			if c.Request.Method == "GET" {
				c.Set(DATA, inValInterface)
			} else {
				err := c.Bind(inValInterface)
				if err != nil {
					var body []byte
					var readBodyErr error
					if c.Request.Method == "POST" {
						body, readBodyErr = ioutil.ReadAll(c.Request.Body)
					} else {
						body = []byte(c.Request.RequestURI)
					}
					doglog.Error("[Warp] data not valid!data=%s,func=%v,err=%v,readBodyErr=%v", string(body), toWrap, err, readBodyErr)
					Return(c, http.StatusBadRequest, "data not valid", err, nil)
					c.Set(SESSION_LOG_LEVEL, "INFO")
					return
				}
				c.Set(DATA, inValInterface)
			}
		} else {
			dataBts, ok := dataBtsObj.([]byte)
			if !ok {
				doglog.Error("[Warp] data not []byte!func=%v,data=%v", toWrap, dataBtsObj)
				Return(c, http.StatusInternalServerError, "data not byte array", nil, nil)
				c.Set(SESSION_LOG_LEVEL, "INFO")
				return
			}
			if dataBts != nil && len(dataBts) > 0 {
				jsonErr := json.Unmarshal(dataBts, inValInterface)
				if jsonErr != nil {
					doglog.Info("[Warp] wrap data from json fail!bts=%s,func=%v,err=%v", string(dataBts), toWrap, jsonErr)
					Return(c, http.StatusInternalServerError, "data type not valid", jsonErr, nil)
					c.Set(SESSION_LOG_LEVEL, "INFO")
					return
				}
			} else {
				if inType.Kind() == reflect.Ptr {
					inVal = reflect.Zero(inType)
					inValInterface = inVal.Interface()
				}
			}
			c.Set(DATA, inValInterface)
		}

		in := make([]reflect.Value, wtNumIn)
		in[0] = reflect.ValueOf(c)
		in[1] = inVal
		out := refToWrap.Call(in)
		if len(out) != 4 {
			doglog.Error("[Warp] return not 4!in=%v,out=%v,func=%v", in, out, toWrap)
			Return(c, http.StatusInternalServerError, "ret not 5!", nil, nil)
			return
		}

		var (
			code    int
			message string
			err     error
			ret     interface{}
		)

		if out[0].CanInterface() {
			code, _ = out[0].Interface().(int)
		} else {
			code = http.StatusInternalServerError
			doglog.Error("[Warp] unparseable code!in=%v,out=%v,func=%v", in, out, toWrap)
		}

		if out[1].CanInterface() {
			message, _ = out[1].Interface().(string)
		} else {
			doglog.Error("[Warp] unparseable message!in=%v,out=%v,func=%v", in, out, toWrap)
		}

		if out[2].CanInterface() {
			err, _ = out[2].Interface().(error)
		} else {
			doglog.Error("[Warp] unparseable err!in=%v,out=%v,func=%v", in, out, toWrap)
		}
		if out[3].CanInterface() {
			ret = out[3].Interface()
		} else {
			doglog.Error("[Warp] unparseable result!in=%v,out=%v,func=%v", in, out, toWrap)
		}

		doglog.Debug("[Warp] wrapped call,in=%v,out=%v,func=%v", in, out, toWrap)
		Return(c, code, message, err, ret)
	}
	return wrapped, nil
}

func Return(c *gin.Context, code int, message string, err error, result interface{}) {
	ret := make(map[string]interface{})
	ret["code"] = code
	ret["result"] = result
	ret["message"] = message

	c.Set(RET, ret)
	c.Set(CODE, code)
	if err != nil {
		c.Set(ERR, err)
	}
}
