/**
 * Copyright 2018 Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"fmt"
	"github.com/chuck1024/dlog"
	"github.com/chuck1024/gd"
	"github.com/chuck1024/gd/server/discovery"
	"time"
)

func main() {
	d := gd.Default()
	c := d.NewRpcClient(time.Duration(500*time.Millisecond), 0)
	// discovery
	var r discovery.DogDiscovery
	r = &discovery.EtcdDiscovery{}
	r.NewDiscovery([]string{"localhost:2379"})
	r.Watch("/root/github/gd/stagging/pool")
	r.Run()
	time.Sleep(100 * time.Millisecond)

	hosts := r.GetNodeInfo("/root/github/gd/stagging/pool")
	for _, v := range hosts {
		dlog.Debug("%s:%d", v.GetIp(), v.GetPort())
	}

	// you can choose one or use load balance algorithm to choose best one.
	// or put all to c.Addr
	for _, v := range hosts {
		if !v.GetOffline() {
			c.AddAddr(fmt.Sprintf("%s:%d", v.GetIp(), v.GetPort()))
		}
	}

	body := []byte("How are you?")

	code, rsp, err := c.DogInvoke(1024, body)
	if err != nil {
		dlog.Error("Error when sending request to server: %s", err)
	}

	dlog.Debug("code=%d, resp=%s", code, string(rsp))
}
