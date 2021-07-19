package main

import (
	"fmt"
	"github.com/chuck1024/gd"
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/service/discovery"
	"time"
)

func main() {
	var i chan struct{}
	c := gd.NewRpcClient(time.Duration(500*time.Millisecond), 0)
	// discovery
	var r discovery.DogDiscovery

	r = &discovery.EtcdDiscovery{}
	if err := r.Start(); err != nil {
		dlog.Error("err:%s", err)
		return
	}

	if err := r.Watch("test", "/root/github/gd/prod/pool"); err != nil {
		dlog.Error("err:%s", err)
		return
	}
	time.Sleep(100 * time.Millisecond)

	hosts := r.GetNodeInfo("test")
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

	// or use rpc protocol
	//rsp, err = c.Invoke(1024, body)
	//if err != nil {
	//t.Logf("Error when sending request to server: %s", err)
	//}

	dlog.Debug("code=%d,resp=%s", code, string(rsp))
	<-i
}
