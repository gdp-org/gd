/**
 * Copyright 2021 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/chuck1024/gd"
	de "github.com/chuck1024/gd/derror"
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/net/dogrpc"
)

type TestReq struct {
	Data string
}

type TestResp struct {
	Ret string
}

func test(req *TestReq) (code uint32, message string, err error, ret *TestResp) {
	dlog.Debug("rpc sever req:%v", req)

	ret = &TestResp{
		Ret: "ok!!!",
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
