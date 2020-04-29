/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/chuck1024/doglog"
	"github.com/chuck1024/godog"
	de "github.com/chuck1024/godog/error"
	"github.com/chuck1024/godog/net/dogrpc"
	"github.com/chuck1024/godog/net/httplib"
	"github.com/chuck1024/godog/server/register"
	"github.com/chuck1024/godog/utils"
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
	doglog.Debug("httpServerTest req:%v", req)

	ret = &TestResp{
		Ret: "ok!!!",
	}

	return http.StatusOK, "ok", nil, ret
}

func HandlerRpcTest(req *TestReq) (code uint32, message string, err error, ret *TestResp) {
	doglog.Debug("rpc sever req:%v", req)

	ret = &TestResp{
		Ret: "ok!!!",
	}

	return uint32(de.RpcSuccess), "ok", nil, ret
}

func Register(e *godog.Engine) {
	// http
	e.HttpServer.DefaultAddHandler("test", HandlerHttpTest)
	e.HttpServer.SetInit(func(g *gin.Engine) error {
		r := g.Group("")
		r.Use(
			httplib.GlFilter(),
			httplib.GroupFilter(),
			httplib.Logger(),
		)

		for k, v := range e.HttpServer.DefaultHandlerMap {
			f, err := httplib.Wrap(v)
			if err != nil {
				return err
			}
			r.POST(k, f)
		}

		return nil
	})

	// Rpc
	e.RpcServer.AddDogHandler(1024, HandlerRpcTest)
	if err := e.RpcServer.DogRpcRegister(); err != nil {
		doglog.Error("DogRpcRegister occur error:%s", err)
		return
	}
	dogrpc.InitFilters([]dogrpc.Filter{&dogrpc.GlFilter{}, &dogrpc.LogFilter{}})
}

func main() {
	d := godog.Default()
	d.InitLog()

	Register(d)

	// register params
	etcdHost, _ := d.Config.Strings("etcdHost")
	root, _ := d.Config.String("root")
	environ, _ := d.Config.String("environ")
	group, _ := d.Config.String("group")
	weight, _ := d.Config.Int("weight")

	// register
	var r register.DogRegister
	r = &register.EtcdRegister{}
	r.NewRegister(etcdHost, root, environ, group, d.Config.BaseConfig.Server.AppName)
	r.Run(utils.GetLocalIP(), d.Config.BaseConfig.Server.RpcPort, uint64(weight))

	err := d.Run()
	if err != nil {
		doglog.Error("Error occurs, error = %s", err.Error())
		return
	}
}

// you can use command to test http service.
// curl -X POST http://127.0.0.1:10240/test -H "Content-Type: application/json" --data '{"Data":"test"}'
