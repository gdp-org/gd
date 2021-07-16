/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package discovery

import (
	"github.com/chuck1024/gd/service"
)

var (
	defaultConf = "conf/conf.ini"
)

type DogDiscovery interface {
	Start() error
	Close()
	Watch(key, node string) error
	WatchMulti(nodes map[string]string) error
	AddNode(key string, info service.NodeInfo)
	DelNode(key string, addr string)
	GetNodeInfo(key string) (nodesInfo []service.NodeInfo)
}
