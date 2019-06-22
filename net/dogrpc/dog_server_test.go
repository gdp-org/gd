/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc_test

import (
	"github.com/chuck1024/godog"
	"github.com/chuck1024/godog/net/dogrpc"
	"testing"
)

func TestDogServer(t *testing.T) {
	d := godog.Default()
	// Tcp
	d.RpcServer = dogrpc.NewDogRpcServer()
	d.RpcServer.AddHandler(1024, func(req []byte) (uint32, []byte) {
		t.Logf("tcp server request: %s", string(req))
		code := uint32(0)
		resp := []byte("Are you ok?")
		return code, resp
	})

	err := d.RpcServer.Run(10241)
	if err != nil {
		t.Logf("Error occurs, error = %s", err.Error())
		return
	}
}
