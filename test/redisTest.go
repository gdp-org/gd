/**
 * Copyright 2018 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/chuck1024/godog"
	"github.com/chuck1024/godog/store/cache"
)

func main() {
	key := "key"
	if err := cache.RedisHandle.Set(key, "value", 10, 0, false, true); err != nil {
		godog.Error("redis set occur error:%s", err)
		return
	}

	value, err := cache.RedisHandle.Get(key)
	if err != nil {
		godog.Error("redis get occur error:%s", err)
		return
	}

	godog.Debug("value:%s", string(value))
}
