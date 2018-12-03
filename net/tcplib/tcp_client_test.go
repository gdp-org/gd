/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib_test

import (
	"github.com/chuck1024/godog"
	"testing"
)

func TestTcpClient(t *testing.T) {
	c := godog.NewTcpClient(500, 0)
	// discovery
	c.AddAddr("127.0.0.1:10240")

	body := []byte("How are you?")

	rsp, err := c.Invoke(1024, body)
	if err != nil {
		t.Logf("Error when sending request to server: %s", err)
	}

	// or use godog protocol
	//rsp, err = c.DogInvoke(1024, body)
	//if err != nil {
	//	t.Logf("Error when sending request to server: %s", err)
	//}

	t.Logf("resp=%s", string(rsp))
}
