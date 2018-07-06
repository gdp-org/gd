/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/xuyu/logging"
	"godog/net/tcplib"
)

func main() {
	c := tcplib.NewClient(500, 0)
	// remember alter addr
	c.AddAddr("10.235.202.118:10241")

	body := []byte("How are you?")

	rsp, err := c.Invoke(1024, body)
	if err != nil {
		logging.Error("Error when sending request to server: %s", err)
	}

	logging.Debug("resp=%s", string(rsp))
}
