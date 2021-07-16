package main

import (
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/service/discovery"
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
	var i chan struct{}
	etcdDis()
	<-i
}
