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

func TestDogClient(t *testing.T) {
	body := &struct {
		Data string
	}{
		Data: "How are you?",
	}

	// use gd protocol
	code, rsp, err := gd.NewRpcClient(500*time.Millisecond, 0).AddAddr(network.GetLocalIP() + ":10241").DogInvoke(1024, body)
	if err != nil {
		t.Logf("Error when sending request to server: %s", err)
	}

	t.Logf("code=%d, resp=%s", code, string(rsp))
}
