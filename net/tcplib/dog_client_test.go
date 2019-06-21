/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib_test

import (
	"github.com/chuck1024/godog"
	"github.com/chuck1024/godog/utils"
	"testing"
)

func TestDogClient(t *testing.T) {
	d := godog.Default()
	c := d.NewTcpClient(500, 0)
	c.AddAddr(utils.GetLocalIP() + ":10241")

	body := []byte("How are you?")

	// use godog protocol
	rsp, err := c.DogInvoke(1024, body)
	if err != nil {
		t.Logf("Error when sending request to server: %s", err)
	}

	t.Logf("resp=%s", string(rsp))
}
