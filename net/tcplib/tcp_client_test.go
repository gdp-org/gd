/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib_test

import (
	"github.com/chuck1024/godog"
	"github.com/chuck1024/godog/utils"
	"testing"
	"time"
)

func TestTcpClient(t *testing.T) {
	d := godog.Default()
	c := d.NewTcpClient(time.Duration(500*time.Millisecond), 0)
	c.AddAddr(utils.GetLocalIP() + ":10241")

	body := []byte("How are you?")

	rsp, err := c.Invoke(1024, body)
	if err != nil {
		t.Logf("Error when sending request to server: %s", err)
	}

	t.Logf("resp=%s", string(rsp))
}
