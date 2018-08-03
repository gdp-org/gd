/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"godog"
	"godog/net/tcplib"
)

func main() {
	c := tcplib.NewClient(500, 0)
	// remember alter addr
	c.AddAddr("127.0.0.1:10241")

	body := []byte("How are you?")

	rsp, err := c.Invoke(1024, body)
	if err != nil {
		godog.Error("Error when sending request to server: %s", err)
	}

	godog.Debug("resp=%s", string(rsp))
}
