/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/chuck1024/gd"
	"github.com/chuck1024/gd/databases/redisdb"
)

func main() {
	defer gd.LogClose()
	t := &redisdb.RedisConfig{
		Addrs: []string{"127.0.0.1:6379"},
	}

	o := &redisdb.RedisPoolClient{
		RedisConfig: t,
	}

	err := o.Start()
	if err != nil {
		gd.Debug("err:%s", err)
	}

	o.Set("test", "ok")
	v, err := o.Get("test")
	if err != nil {
		gd.Debug("err:%s", err)
	}
	gd.Debug("%s", v)
}
