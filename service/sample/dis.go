package main

import (
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/service/discovery"
	"time"
)

func etcdDis() {
	var r discovery.DogDiscovery
	var i chan struct{}

	r = &discovery.EtcdDiscovery{}
	if err := r.Start(); err != nil {
		dlog.Error("err:%s", err)
		return
	}
	defer r.Close()

	r.Watch("test", "/root/github/gd/prod/pool")
	time.Sleep(100 * time.Millisecond)

	n1 := r.GetNodeInfo("test")
	for _, v := range n1 {
		dlog.Info("%s:%d", v.GetIp(), v.GetPort())
	}
	<-i
}

func zkDis() {
	var r discovery.DogDiscovery
	var i chan struct{}
	r = &discovery.ZkDiscovery{}
	if err := r.Start(); err != nil {
		dlog.Error("err:%s", err)
		return
	}
	defer r.Close()

	r.Watch("test","/root/github/gd/prod/pool")

	time.Sleep(100 * time.Millisecond)
	n1 := r.GetNodeInfo("test")
	for _, v := range n1 {
		dlog.Info("%s:%d", v.GetIp(), v.GetPort())
	}
	<-i
}

func main() {
	etcdDis()
}
