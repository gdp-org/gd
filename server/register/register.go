/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package register

import "time"

var (
	DefaultHeartBeat  uint64 = 10
	DefaultRetryTimes        = 3
)

// register server
type DogRegister interface {
	NewRegister(hosts []string, root, environ, group, service string)
	SetRootNode(node string) error
	GetRootNode() (root string)
	SetHeartBeat(heartBeat time.Duration)
	SetOffline(offline bool)
	Run(ip string, port int, weight uint64) error
	Close()
}
