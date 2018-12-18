/**
 * Copyright 2018 Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"fmt"
	"github.com/chuck1024/godog"
	"github.com/chuck1024/godog/server/discovery"
	"time"
)

func main() {
	c := godog.NewTcpClient(500, 0)
	// discovery
	var r discovery.DogDiscovery
	r = &discovery.EtcdDiscovery{}
	r.NewDiscovery([]string{"localhost:2379"})
	r.Watch("/root/github/godog/stagging/pool")
	r.Run()
	time.Sleep(100 * time.Millisecond)

	hosts := r.GetNodeInfo("/root/github/godog/stagging/pool")
	for _, v := range hosts {
		godog.Debug("%s:%d", v.GetIp(), v.GetPort())
	}

	// you can choose one or use load balance algorithm to choose best one.
	// or put all to c.Addr
	for _, v := range hosts {
		if !v.GetOffline() {
			c.AddAddr(fmt.Sprintf("%s:%d", v.GetIp(), v.GetPort()))
		}
	}

	body := []byte("How are you?")

	rsp, err := c.Invoke(1024, body)
	if err != nil {
		godog.Error("Error when sending request to server: %s", err)
	}

	godog.Debug("resp=%s", string(rsp))
}
