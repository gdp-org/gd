/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package discovery

import (
	"encoding/json"
	"fmt"
	"github.com/chuck1024/godog/server"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/xuyu/logging"
	"sync"
	"time"
)

type ZkNode struct {
	node      string
	nodesInfo []server.NodeInfo
	stopChan  chan struct{}
	client    *zk.Conn
}

// Encapsulates the zookeeper discovery
type ZkDiscovery struct {
	dns      []string          // zookeeper host
	nodes    map[string]ZkNode // watch node
	exitChan chan struct{}
	running  bool
	lock     *sync.Mutex
}

func (z *ZkDiscovery) NewDiscovery(dns []string) {
	z.lock = new(sync.Mutex)
	z.nodes = make(map[string]ZkNode)
	z.dns = dns
	z.running = false
}

func (z *ZkDiscovery) Watch(node string) error {
	if node[0] != '/' {
		node = "/" + node
	}

	zkNode := ZkNode{
		node:     node,
		stopChan: make(chan struct{}, 1),
	}

	conn, _, err := zk.Connect(z.dns, time.Second*5, zk.WithLogInfo(false))
	if err != nil {
		logging.Error("[Watch] zk connect occur error:%s", err)
		return err
	}

	zkNode.client = conn

	z.lock.Lock()
	defer z.lock.Unlock()
	z.nodes[node] = zkNode
	if z.running {
		go z.watchNode(zkNode)
	}

	return nil
}

func (z *ZkDiscovery) WatchMulti(nodes []string) error {
	for _, node := range nodes {
		err := z.Watch(node)
		if err != nil {
			return err
		}
	}

	return nil
}

func (z *ZkDiscovery) AddNode(node string, info *server.NodeInfo) {
	zkNode := z.nodes[node]
	z.lock.Lock()
	defer z.lock.Unlock()
	zkNode.nodesInfo = append(zkNode.nodesInfo, *info)
	z.nodes[node] = zkNode
	return
}

func (z *ZkDiscovery) DelNode(node string, key string) {
	zkNode := z.nodes[node]
	for k, v := range zkNode.nodesInfo {
		if v.GetIp()+fmt.Sprintf(":%d", v.GetPort()) == key {
			z.lock.Lock()
			zkNode.nodesInfo = append(zkNode.nodesInfo[:k], zkNode.nodesInfo[k+1:]...)
			z.nodes[node] = zkNode
			z.lock.Unlock()
			break
		}
	}
}

func (z *ZkDiscovery) unMsgNodeInfo(data []byte) *server.NodeInfo {
	var info server.NodeInfo
	info = &server.DefaultNodeInfo{}
	err := json.Unmarshal([]byte(data), info)
	if err != nil {
		logging.Error("[GetNodeInfo] json unmarshal occur error:%s", err)
		return nil
	}

	return &info
}

func (z *ZkDiscovery) GetNodeInfo(node string) []server.NodeInfo {
	return z.nodes[node].nodesInfo
}

func (z *ZkDiscovery) watchNode(node ZkNode) {
	children, _, err := z.nodes[node.node].client.Children(node.node)
	if err != nil {
		logging.Error("[watchNode] Children occur error:%s", err)
		return
	}

	for _, v := range children {
		data, _, err := z.nodes[node.node].client.Get(node.node + "/" + v)
		if err != nil {
			logging.Error("[watchNode] get occur error:%s", err)
			return
		}
		zkNode := z.unMsgNodeInfo(data)
		z.AddNode(node.node, zkNode)
	}

	for {
		_, _, childCh, err := z.nodes[node.node].client.ChildrenW(node.node)
		if err != nil {
			logging.Error("[watchNode] watch childrenW occur error:%s", err)
			return
		}

		select {
		case childEvent := <-childCh:
			if childEvent.Type == zk.EventNodeChildrenChanged {
				children, _, err := z.nodes[node.node].client.Children(node.node)
				if err != nil {
					logging.Error("[watchNode] Children occur error:%s", err)
					return
				}

				zkNode := z.nodes[node.node]
				zkNode.nodesInfo = nil
				z.nodes[node.node] = zkNode
				for _, v := range children {
					data, _, err := z.nodes[node.node].client.Get(node.node + "/" + v)
					if err != nil {
						logging.Error("[watchNode] get occur error:%s", err)
						return
					}
					zkNode := z.unMsgNodeInfo(data)
					z.AddNode(node.node, zkNode)
				}
			}
		}
	}

}

func (z *ZkDiscovery) Run() error {
	if z.running {
		return fmt.Errorf("zk discovery is already running")
	}

	z.lock.Lock()
	defer z.lock.Unlock()
	z.running = true
	z.exitChan = make(chan struct{}, MaxNodeNum)

	for _, nodes := range z.nodes {
		go z.watchNode(nodes)
	}
	return nil
}

func (z *ZkDiscovery) Close() error {
	for _, node := range z.nodes {
		close(node.stopChan)
	}

	length := len(z.nodes)
	for i := 0; i < length; i++ {
		<-z.exitChan
	}

	for _, node := range z.nodes {
		close(node.stopChan)
	}

	if z.exitChan != nil {
		close(z.exitChan)
	}

	return nil
}
