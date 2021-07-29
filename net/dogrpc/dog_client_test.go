/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc_test

import (
	"github.com/chuck1024/gd"
	"github.com/chuck1024/gd/utls/network"
	"testing"
	"time"
)

func TestDogClient(t *testing.T) {
	c := gd.NewRpcClient(time.Duration(500*time.Millisecond), 0, false)
	c.AddAddr(network.GetLocalIP() + ":10241")

	body := &struct {
		Data string
	}{
		Data: "How are you?",
	}

	// use gd protocol
	code, rsp, err := c.DogInvoke(1024, body)
	if err != nil {
		t.Logf("Error when sending request to server: %s", err)
	}

	t.Logf("code=%d, resp=%s", code, string(rsp))
}
