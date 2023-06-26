/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc_test

import (
	de "github.com/gdp-org/gd/derror"
	"github.com/gdp-org/gd/dlog"
	"github.com/gdp-org/gd/net/dogrpc"
	"testing"
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

func TestDogServer(t *testing.T) {
	var i chan struct{}
	d := dogrpc.NewDogRpcServer()
	d.Addr = 10241
	// Rpc
	d.AddDogHandler(1024, test)
	if err := d.DogRpcRegister(); err != nil {
		t.Logf("DogRpcRegister occur error:%s", err)
		return
	}

	err := d.Start()
	if err != nil {
		t.Logf("Error occurs, error = %s", err.Error())
		return
	}
	<-i
}
