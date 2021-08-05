package main

import (
	"fmt"
	"github.com/gdp-org/gd"
	"github.com/gdp-org/gd/service/discovery"
	"time"
)

func main() {
	defer gd.LogClose()
	c := gd.NewRpcClient(time.Duration(500*time.Millisecond), 0)
	// discovery
	var r discovery.DogDiscovery

	r = &discovery.EtcdDiscovery{}
	if err := r.Start(); err != nil {
		gd.Error("err:%s", err)
		return
	}

	if err := r.Watch("test", "/root/github/gd/prod/pool"); err != nil {
		gd.Error("err:%s", err)
		return
	}
	time.Sleep(100 * time.Millisecond)

	hosts := r.GetNodeInfo("test")
	for _, v := range hosts {
		gd.Debug("%s:%d", v.GetIp(), v.GetPort())
	}

	// you can choose one or use load balance algorithm to choose best one.
	// or put all to c.Addr
	for _, v := range hosts {
		if !v.GetOffline() {
			c.AddAddr(fmt.Sprintf("%s:%d", v.GetIp(), v.GetPort()))
		}
	}

	body := &struct {
		Data string
	}{
		Data: "How are you?",
	}

	code, rsp, err := c.DogInvoke(1024, body)
	if err != nil {
		gd.Error("Error when sending request to server: %s", err)
	}

	// or use rpc protocol
	//rsp, err = c.Invoke(1024, body)
	//if err != nil {
	//t.Logf("Error when sending request to server: %s", err)
	//}

	gd.Debug("code=%d,resp=%s", code, string(rsp))
}
