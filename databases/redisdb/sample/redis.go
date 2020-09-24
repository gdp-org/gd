/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/chuck1024/gd/databases/redisdb"
	"github.com/chuck1024/gd/dlog"
)

func main() {
	var i chan struct{}
	t := &redisdb.RedisConfig{
		Addrs: []string{"127.0.0.1:6379"},
	}
	o, err := redisdb.NewRedisPools(t)
	if err != nil {
		dlog.Debug("err:%s", err)
	}
	o.Set("test", "ok")
	v, err := o.Get("test")
	if err != nil {
		dlog.Debug("err:%s", err)
	}
	dlog.Debug("%s", v)
	<-i
}
