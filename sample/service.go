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
	"net/http"
)

func HandlerHttpTest(w http.ResponseWriter, r *http.Request) {
	doglog.Debug("connected : %s", r.RemoteAddr)
	w.Write([]byte("test success!!!"))
}

func HandlerTcpTest(req []byte) (uint32, []byte) {
	doglog.Debug("tcp server request: %s", string(req))
	code := uint32(200)
	resp := []byte("Are you ok?")
	return code, resp
}

func main() {
	// Http
	godog.AppHttp.AddHttpHandler("/test", HandlerHttpTest)

	// default tcp server, you can choose godog tcp server
	//godog.AppTcp = tcplib.AppDog

	// Tcp
	godog.AppTcp.AddTcpHandler(1024, HandlerTcpTest)

	// register params
	etcdHost, _ := godog.AppConfig.Strings("etcdHost")
	root, _ := godog.AppConfig.String("root")
	environ, _ := godog.AppConfig.String("environ")
	group, _ := godog.AppConfig.String("group")
	weight, _ := godog.AppConfig.Int("weight")

	// register
	var r register.DogRegister
	r = &register.EtcdRegister{}
	r.NewRegister(etcdHost, root, environ, group, godog.AppConfig.BaseConfig.Server.AppName)
	r.Run(utils.GetLocalIP(), godog.AppConfig.BaseConfig.Server.TcpPort, uint64(weight))

	err := godog.Run()
	if err != nil {
		doglog.Error("Error occurs, error = %s", err.Error())
		return
	}
}

// you can use command to test http service.
// curl http://127.0.0.1:10240/test
