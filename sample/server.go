/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/chuck1024/gd"
	de "github.com/chuck1024/gd/derror"
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/net/dhttp"
	"github.com/chuck1024/gd/net/dogrpc"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TestReq struct {
	Data string
}

type TestResp struct {
	Ret string
}

func HandlerHttpTest(c *gin.Context, req *TestReq) (code int, message string, err error, ret *TestResp) {
	dlog.Debug("httpServerTest req:%v", req)

	ret = &TestResp{
		Ret: "ok!!!",
	}

	return http.StatusOK, "ok", nil, ret
}

func HandlerRpcTest(req *TestReq) (code uint32, message string, err error, ret *TestResp) {
	dlog.Debug("rpc sever req:%v", req)

	ret = &TestResp{
		Ret: "ok!!!",
	}

	return uint32(de.RpcSuccess), "ok", nil, ret
}

func Register(e *gd.Engine) {
	// http
	e.HttpServer.SetInit(func(g *gin.Engine) error {
		r := g.Group("")
		r.Use(
			dhttp.GlFilter(),
			dhttp.StatFilter(),
			dhttp.GroupFilter(),
			dhttp.Logger("sample"),
		)

		e.HttpServer.POST(r, "test", HandlerHttpTest)

		if err := e.HttpServer.CheckHandle(); err != nil {
			return err
		}

		return nil
	})

	// Rpc
	e.RpcServer.AddDogHandler(1024, HandlerRpcTest)
	if err := e.RpcServer.DogRpcRegister(); err != nil {
		dlog.Error("DogRpcRegister occur error:%s", err)
		return
	}
	dogrpc.InitFilters([]dogrpc.Filter{&dogrpc.GlFilter{}, &dogrpc.LogFilter{}})
}

func main() {
	d := gd.Default()

	Register(d)

	err := d.Run()
	if err != nil {
		dlog.Error("Error occurs, error = %s", err.Error())
		return
	}
}

// you can use command to test http service.
// curl -X POST http://127.0.0.1:10240/test -H "Content-Type: application/json" --data '{"Data":"test"}'
