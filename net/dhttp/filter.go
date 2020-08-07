/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package dhttp

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/chuck1024/dlog"
	"github.com/chuck1024/gd/utls"
	"github.com/chuck1024/gd/utls/network"
	"github.com/chuck1024/gl"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"time"
)

// example: group filter
func GroupFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		ret, _ := ParseRet(c)
		httpStatusInterface, _ := c.Get(Code)
		httpStatus := httpStatusInterface.(int)
		c.JSON(httpStatus, ret)
	}
}

// example: use gl
func GlFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		gl.Init()
		gl.SetLogger(dlog.Global)
		defer gl.Close()
		c.Next()
	}
}

// example: log middle handle
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		st := time.Now()

		// traceId
		traceId := c.Query("traceId")
		if traceId != "" {
			gl.Set(gl.LogId, traceId)
		} else {
			traceId = utls.TraceId()
			c.Set(TraceID, traceId)
			gl.Set(gl.LogId, traceId)
		}

		realIp, _ := network.GetRealIP(c.Request)
		c.Set(RemoteIP, realIp)
		gl.Set(gl.ClientIp, realIp)

		c.Next()

		uri := c.Request.RequestURI
		uriSplits := strings.Split(uri, "?")
		path := uri
		if len(uriSplits) > 0 {
			path = uriSplits[0]
		}

		costDu := time.Now().Sub(st)
		cost := costDu / time.Millisecond

		data, ok := c.Get(Data)
		if !ok {
			dataRaw, ok := c.Get(DataRaw)
			if ok {
				paramsBts, ok := dataRaw.([]byte)
				if !ok {
					data = fmt.Sprintf("%v", dataRaw)
				} else {
					data = string(paramsBts)
				}
			}
		}

		ret, _ := c.Get(Ret)
		httpStatusInterface, _ := c.Get(Code)
		httpStatus := httpStatusInterface.(int)

		handleErr, _ := c.Get(Err)
		errStr := ""
		handleErrErr, ok := handleErr.(error)
		if ok {
			if handleErrErr != nil {
				errStr = handleErrErr.Error()
			}
		} else {
			if handleErr != nil {
				errStr = fmt.Sprintf("%v", handleErr)
			}
		}

		message := map[string]interface{}{
			"httpStatus": httpStatus,
			"cost":       strconv.FormatInt(int64(cost), 10) + "ms",
			"err":        errStr,
		}

		dataByte, err := json.Marshal(data)
		if err != nil {
			dlog.Error("data cant transfer to json ?! data is %v", data)
			message["data"] = data
		} else {
			dataJson, _ := simplejson.NewJson(dataByte)
			message["data"] = dataJson
		}
		retByte, err := json.Marshal(ret)
		if err != nil {
			dlog.Error("ret cant transfer to json ?! ret is %v", ret)
			message["ret"] = ret
		} else {
			retStr, _ := simplejson.NewJson(retByte)
			message["ret"] = retStr
		}

		mj, jsonErr := utls.Marshal(message)
		if jsonErr != nil {
			dlog.Error("json marshal occur error:%v", jsonErr)
		}

		if cost > 500 {
			dlog.WarnT("SESSION_SLOW", fmt.Sprintf("%s %s", path, string(mj)))
			return
		}
		dlog.InfoT("SESSION", fmt.Sprintf("%s %s", path, string(mj)))
	}
}
