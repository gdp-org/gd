/**
 * Copyright 2021 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/chuck1024/gd"
	"github.com/chuck1024/gd/utls/network"
	"time"
)

func main() {
	defer gd.LogClose()
	c := gd.NewRpcClientTls(time.Duration(500*time.Millisecond), 0, true)
	c.AddAddr(network.GetLocalIP() + ":10241")

	body := &struct {
		Data string
	}{
		Data: "How are you?",
	}

	// use gd protocol
	code, rsp, err := c.DogInvoke(1024, body)
	if err != nil {
		gd.Error("Error when sending request to server: %s", err)
	}

	gd.Info("code=%d, resp=%s", code, string(rsp))
}
