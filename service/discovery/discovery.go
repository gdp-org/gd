/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package discovery

import (
	"github.com/chuck1024/gd/service"
)

const (
	MaxNodeNum = 128
)

type DogDiscovery interface {
	NewDiscovery(dns []string)
	Watch(node string) error
	WatchMulti(nodes []string) error
	AddNode(node string, info *service.NodeInfo)
	DelNode(node string, key string)
	GetNodeInfo(node string) (nodesInfo []service.NodeInfo)
	Run() error
	Close() error
}
