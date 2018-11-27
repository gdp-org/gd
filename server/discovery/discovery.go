/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package discovery

import "github.com/chuck1024/godog/server"

const (
	MaxNodeNum = 128
)

type DogDiscovery interface {
	NewDiscovery(dns []string)
	Watch(node string) error
	WatchMulti(nodes []string) error
	AddNode(node string, info *server.NodeInfo)
	DelNode(node string, key string)
	GetNodeInfo(node string) (nodesInfo []server.NodeInfo)
	Run() error
	Close() error
}
