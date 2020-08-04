/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chuck1024/dlog"
	"github.com/chuck1024/gd/server"
	"github.com/etcd-io/etcd/clientv3"
	"sync"
	"time"
)

type EtcdNode struct {
	node      string
	nodesInfo []server.NodeInfo
	stopChan  chan struct{}
	client    *clientv3.Client
}

// Encapsulates the etcd discovery
type EtcdDiscovery struct {
	dns      []string            //etcd host
	nodes    map[string]EtcdNode // watch node
	exitChan chan struct{}
	running  bool
	lock     *sync.Mutex
}

func (e *EtcdDiscovery) NewDiscovery(dns []string) {
	e.lock = new(sync.Mutex)
	e.nodes = make(map[string]EtcdNode)
	e.dns = dns
	e.running = false
}

func (e *EtcdDiscovery) Watch(node string) error {
	if node[0] != '/' {
		node = "/" + node
	}

	etcdNode := EtcdNode{
		node:     node,
		stopChan: make(chan struct{}, 1),
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   e.dns,
		DialTimeout: time.Second,
	})
	if err != nil {
		dlog.Error("watch new client occur error:%s", err)
		return err
	}

	etcdNode.client = cli

	e.lock.Lock()
	defer e.lock.Unlock()
	e.nodes[node] = etcdNode
	if e.running {
		go e.watchNode(etcdNode)
	}

	return nil
}

func (e *EtcdDiscovery) WatchMulti(nodes []string) error {
	for _, node := range nodes {
		err := e.Watch(node)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *EtcdDiscovery) AddNode(node string, info *server.NodeInfo) {
	etcdNode := e.nodes[node]
	e.lock.Lock()
	defer e.lock.Unlock()
	etcdNode.nodesInfo = append(etcdNode.nodesInfo, *info)
	e.nodes[node] = etcdNode
	return
}

func (e *EtcdDiscovery) DelNode(node string, key string) {
	etcdNode := e.nodes[node]
	for k, v := range etcdNode.nodesInfo {
		if v.GetIp()+fmt.Sprintf(":%d", v.GetPort()) == key {
			e.lock.Lock()
			etcdNode.nodesInfo = append(etcdNode.nodesInfo[:k], etcdNode.nodesInfo[k+1:]...)
			e.nodes[node] = etcdNode
			e.lock.Unlock()
			break
		}
	}
}

func (e *EtcdDiscovery) unMsgNodeInfo(data []byte) *server.NodeInfo {
	var info server.NodeInfo
	info = &server.DefaultNodeInfo{}
	err := json.Unmarshal([]byte(data), info)
	if err != nil {
		dlog.Error("GetNodeInfo json unmarshal occur error:%s", err)
		return nil
	}

	return &info
}

func (e *EtcdDiscovery) GetNodeInfo(node string) []server.NodeInfo {
	return e.nodes[node].nodesInfo
}

func (e *EtcdDiscovery) watchNode(node EtcdNode) {
	resp, err := e.nodes[node.node].client.Get(context.TODO(), node.node, clientv3.WithPrefix())
	if err != nil {
		dlog.Error("watch node get node[%s] children", node.node)
		return
	}

	if resp.Count != 0 {
		for _, ev := range resp.Kvs {
			info := e.unMsgNodeInfo(ev.Value)
			e.AddNode(node.node, info)
		}
	}

	rch := e.nodes[node.node].client.Watch(context.Background(), node.node, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				fmt.Printf("[%s] %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				info := e.unMsgNodeInfo(ev.Kv.Value)
				e.AddNode(node.node, info)
			case clientv3.EventTypeDelete:
				fmt.Printf("[%s] %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				e.DelNode(node.node, string(ev.Kv.Key))
			}
		}
	}
}

func (e *EtcdDiscovery) Run() error {
	if e.running {
		return fmt.Errorf("etcd discovery is already running")
	}

	e.lock.Lock()
	defer e.lock.Unlock()
	e.running = true
	e.exitChan = make(chan struct{}, MaxNodeNum)

	for _, nodes := range e.nodes {
		go e.watchNode(nodes)
	}
	return nil
}

func (e *EtcdDiscovery) Close() error {
	for _, node := range e.nodes {
		close(node.stopChan)
	}

	length := len(e.nodes)
	for i := 0; i < length; i++ {
		<-e.exitChan
	}

	for _, node := range e.nodes {
		close(node.stopChan)
	}

	if e.exitChan != nil {
		close(e.exitChan)
	}

	return nil
}
