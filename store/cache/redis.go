/**
 * Copyright 2018 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package cache

import (
	"github.com/xuyu/goredis"
	"github.com/xuyu/logging"
	"godog/config"
)

var (
	RedisHandle *goredis.Redis
)

func init() {
	url, err := config.AppConfig.String("redis")
	if err != nil {
		logging.Warning("[init] get config redis url occur error: ", err)
		return
	}

	RedisHandle, err = goredis.DialURL(url)
	if err != nil {
		logging.Error("[InitRedis] redis init fail: %s, %s", url, err.Error())
		panic("[InitRedis] InitRedis failed")
		return
	}

	logging.Info("[InitRedis] redis conn ok: %s", url)
}
