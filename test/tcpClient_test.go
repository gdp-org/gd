/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/xuyu/logging"
	"godog/net/tcplib"
	"testing"
)

func TestClient(t *testing.T) {
	c := tcplib.NewClient(500, 0)
	// remember alter addr
	c.AddAddr("192.168.1.107:10241")

	body := []byte("test success")

	rsp, err := c.Invoke(1024, body)
	if err != nil {
		logging.Error("Error when sending request to server: %s", err)
	}

	logging.Debug("resp=%s", string(rsp))
}
