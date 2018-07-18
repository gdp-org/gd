/**
 * Copyright 2018 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/xuyu/logging"
	"godog/store/cache"
)

func main() {
	key := "key"
	if err := cache.RedisHandle.Set(key, "value", 10, 0, false, true); err != nil {
		logging.Error("redis set occur error:%s", err)
		return
	}
}
