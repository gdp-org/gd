/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc_test

import (
	"github.com/chuck1024/gd/net/dogrpc"
	"testing"
)

func TestRpcServer(t *testing.T) {
	d := dogrpc.NewRpcServer()
	// Tcp
	d.AddHandler(1024, func(req []byte) (uint32, []byte) {
		t.Logf("rpc server request: %s", string(req))
		code := uint32(0)
		resp := []byte("Are you ok?")
		return code, resp
	})

	err := d.Start()
	if err != nil {
		t.Logf("Error occurs, derror = %s", err.Error())
		return
	}
}
