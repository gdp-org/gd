/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main_test

import (
	"fmt"
	"github.com/chuck1024/godog"
	"github.com/chuck1024/godog/server/discovery"
	"testing"
	"time"
)

func TestTcpClient(t *testing.T) {
	c := godog.NewTcpClient(500, 0)
	// discovery
	var r discovery.DogDiscovery
	r = &discovery.EtcdDiscovery{}
	r.NewDiscovery([]string{"localhost:2379"})
	r.Watch("/root/github/godog/stagging/pool")
	r.Run()
	time.Sleep(100*time.Millisecond)

	hosts := r.GetNodeInfo("/root/github/godog/stagging/pool")
	for _,v := range hosts {
		t.Logf("%s:%d",v.GetIp(),v.GetPort())
	}

	// you can choose one
	c.AddAddr(hosts[0].GetIp()+":"+fmt.Sprintf("%d",hosts[0].GetPort()))

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
