/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/chuck1024/doglog"
	"github.com/chuck1024/godog"
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

func HandlerTcpTest(req []byte) (uint32, []byte) {
	doglog.Debug("tcp server request: %s", string(req))
	code := uint32(200)
	resp := []byte("Are you ok?")
	return code, resp
}

func main() {
	d := godog.Default()
	// Http
	d.HttpServer.DefaultAddHandler("test", HandlerHttpTest)
	d.HttpServer.DefaultRegister()

	// default tcp server, you can choose godog tcp server
	//d.TcpServer = tcplib.NewDogTcpServer()

	// Tcp
	d.TcpServer.AddTcpHandler(1024, HandlerTcpTest)

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
	r.Run(utils.GetLocalIP(), d.Config.BaseConfig.Server.TcpPort, uint64(weight))

	err := d.Run()
	if err != nil {
		doglog.Error("Error occurs, error = %s", err.Error())
		return
	}
}

// you can use command to test http service.
// curl http://127.0.0.1:10240/test
