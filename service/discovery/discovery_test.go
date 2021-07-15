/**
 * Copyright 2018 Author. All rights reserved.
 * Author: Chuck1024
 */

package discovery_test

import (
	"github.com/chuck1024/gd/service/discovery"
	"testing"
	"time"
)

func TestDiscEtcd(t *testing.T) {
	var r discovery.DogDiscovery
	r = &discovery.EtcdDiscovery{}
	r.NewDiscovery([]string{"localhost:2379"})
	r.Watch("/root/github/gd/stagging/pool")
	r.Run()
	time.Sleep(100 * time.Millisecond)

	n1 := r.GetNodeInfo("/root/github/gd/stagging/pool")
	for _, v := range n1 {
		t.Logf("%s:%d", v.GetIp(), v.GetPort())
	}

	time.Sleep(10 * time.Second)
}

func TestDiscZk(t *testing.T) {
	var r discovery.DogDiscovery
	r = &discovery.ZkDiscovery{}
	r.NewDiscovery([]string{"localhost:2181"})
	r.Watch("/root/gd/test/stagging/pool")
	r.Run()
	time.Sleep(100 * time.Millisecond)
	n1 := r.GetNodeInfo("/root/gd/test/stagging/pool")
	for _, v := range n1 {
		t.Logf("%s:%d", v.GetIp(), v.GetPort())
	}
	time.Sleep(10 * time.Second)
}
