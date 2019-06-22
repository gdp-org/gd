/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc_test

import (
	"github.com/chuck1024/godog/net/dogrpc"
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

	err := d.Run(10241)
	if err != nil {
		t.Logf("Error occurs, error = %s", err.Error())
		return
	}
}
