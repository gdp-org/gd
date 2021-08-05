package main

import (
	"github.com/gdp-org/gd/dlog"
	"github.com/gdp-org/gd/service/discovery"
	"time"
)

func etcdDis() {
	var r discovery.DogDiscovery

	r = &discovery.EtcdDiscovery{}
	if err := r.Start(); err != nil {
		dlog.Error("err:%s", err)
		return
	}

	r.Watch("test", "/root/github/gd/prod/pool")
	time.Sleep(100 * time.Millisecond)

	n1 := r.GetNodeInfo("test")
	for _, v := range n1 {
		dlog.Info("%s:%d", v.GetIp(), v.GetPort())
	}
}

func zkDis() {
	var r discovery.DogDiscovery
	r = &discovery.ZkDiscovery{}
	if err := r.Start(); err != nil {
		dlog.Error("err:%s", err)
		return
	}

	r.Watch("test", "/root/github/gd/prod/pool")

	time.Sleep(100 * time.Millisecond)
	n1 := r.GetNodeInfo("test")
	for _, v := range n1 {
		dlog.Info("%s:%d", v.GetIp(), v.GetPort())
	}
}

func main() {
	defer dlog.Close()
	etcdDis()
}
