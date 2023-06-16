/**
 * Copyright 2021 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/gdp-org/gd"
	de "github.com/gdp-org/gd/derror"
	"github.com/gdp-org/gd/net/dogrpc"
)

type TestReq struct {
	Data string
}

type TestResp struct {
	Ret string
}

func test(req *TestReq) (code uint32, message string, err error, ret *TestResp) {
	gd.Info("rpc sever req:%v", req)

	ret = &TestResp{
		Ret: req.Data,
	}

	return uint32(de.RpcSuccess), "ok", nil, ret
}

func main() {
	var i chan struct{}
	defer gd.LogClose()
	d := dogrpc.NewDogRpcServer()
	d.Addr = 10241
	d.UseTls = true
	// Rpc
	d.AddDogHandler(1024, test)
	if err := d.DogRpcRegister(); err != nil {
		gd.Error("DogRpcRegister occur error:%s", err)
		return
	}

	err := d.Start()
	if err != nil {
		gd.Error("Error occurs, error = %s", err.Error())
		return
	}
	<-i
}
