/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc_test

import (
	"github.com/gdp-org/gd"
	"github.com/gdp-org/gd/utls/network"
	"testing"
	"time"
)

func TestRpcClient(t *testing.T) {
	body := []byte("How are you?")

	code, rsp, err := gd.NewRpcClient(500*time.Millisecond, 0).AddAddr(network.GetLocalIP()+":10241").Invoke(1024, body)
	if err != nil {
		t.Logf("Error when sending request to server: %s", err)
	}

	t.Logf("code=%d, resp=%s", code, string(rsp))
}
