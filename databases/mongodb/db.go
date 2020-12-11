/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package mongodb

const (
	DefaultMaxActive    = 500
	DefaultMaxIdle      = 8
	DefaultIdleTimeout  = 300
	DefaultRetryTimes   = 3
	DefaultConnTimeout  = 400
	DefaultReadTimeout  = 700
	DefaultWriteTimeout = 500

	MongoCommonCostMax   = 20
	MongoCmd             = "mongo_cmd_%v"
	MongoCmdSlowCount    = "mongo_%v_slow_count"
	MongoNormalSlowCount = "mongo_common_slow_count"

	glMongoCall     = "mongo_call"
	glMongoCost     = "mongo_cost"
	glMongoCallFail = "mongo_call_fail"

	defaultConf = "conf/conf.ini"
)

type MongoConfig struct {
	Hosts         []string
	UserName      string
	Password      string
	Database      string
	ConnTimeoutMs int64
	WTimeoutMs    int64
}
