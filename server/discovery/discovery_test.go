/**
 * Copyright 2018 Author. All rights reserved.
 * Author: Chuck1024
 */

package discovery

import (
	"testing"
	"time"
)

func TestDiscEtcd(t *testing.T){
	var r DogDiscovery
	r = &EtcdDiscovery{}
	r.NewDiscovery([]string{"localhost:2379"})
	r.Watch("/root/github/godog/stagging/pool")
	r.Run()
	time.Sleep(100*time.Millisecond)

	n1 := r.GetNodeInfo("/root/github/godog/stagging/pool")
	for _,v := range n1 {
		t.Logf("%s:%d",v.GetIp(),v.GetPort())
	}

	time.Sleep(10*time.Second)
}

func TestDiscZk(t *testing.T){
	var r DogDiscovery
	r = &ZkDiscovery{}
	r.NewDiscovery([]string{"localhost:2181"})
	r.Watch("/root/godog/test/stagging/pool")
	r.Run()
	time.Sleep(100*time.Millisecond)
	n1 := r.GetNodeInfo("/root/godog/test/stagging/pool")
	for _,v := range n1 {
		t.Logf("%s:%d",v.GetIp(),v.GetPort())
	}
	time.Sleep(10*time.Second)
}
