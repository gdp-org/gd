/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package register

import "time"

var (
	DefaultHeartBeat  uint64 = 10
	DefaultRetryTimes        = 3
	defaultConf              = "conf/conf.ini"
)

// register server
type DogRegister interface {
	Start() error
	Close()
	SetRootNode(node string) error
	GetRootNode() (root string)
	SetHeartBeat(heartBeat time.Duration)
	SetOffline(offline bool)
}
