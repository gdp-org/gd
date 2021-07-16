/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package discovery

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/service"
	"github.com/samuel/go-zookeeper/zk"
	"gopkg.in/ini.v1"
	"strings"
	"sync"
	"time"
)

type ZkNode struct {
	key       string
	path      string
	nodesInfo []service.NodeInfo
	stopChan  chan struct{}
	client    *zk.Conn
}

type ZkConfig struct {
	host []string // zk server host
}

// Encapsulates the zookeeper discovery
type ZkDiscovery struct {
	ZkConfig   *ZkConfig `inject:"zkConfig" canNil:"true"`
	ZkConf     *ini.File `inject:"zkConf" canNil:"true"`
	ZkConfPath string    `inject:"zkConfPath" canNil:"true"`

	nodes   sync.Map
	running bool

	startOnce sync.Once
	closeOnce sync.Once
}

func (z *ZkDiscovery) Start() error {
	var err error
	z.startOnce.Do(func() {
		if z.ZkConfig != nil {
			err = z.initWithZkConfig(z.ZkConfig)
		} else if z.ZkConf != nil {
			err = z.initZk(z.ZkConf)
		} else {
			if z.ZkConfPath == "" {
				z.ZkConfPath = defaultConf
			}

			err = z.initObjForZk(z.ZkConfPath)
		}
	})
	return err
}

func (z *ZkDiscovery) Close() {
	z.nodes.Range(func(key, value interface{}) bool {
		close(value.(ZkNode).stopChan)
		return true
	})
	return
}

func (z *ZkDiscovery) initObjForZk(filePath string) error {
	zkConfRealPath := filePath
	if zkConfRealPath == "" {
		return errors.New("zkConf not set")
	}

	if !strings.HasSuffix(zkConfRealPath, ".ini") {
		return errors.New("zkConf not an ini file")
	}

	zkConf, err := ini.Load(zkConfRealPath)
	if err != nil {
		return err
	}

	if err = z.initZk(zkConf); err != nil {
		return err
	}
	return nil
}

func (z *ZkDiscovery) initZk(f *ini.File) error {
	c := f.Section("DisRes")
	hosts := c.Key("zkHost").Strings(",")
	config := &ZkConfig{
		host: hosts,
	}

	return z.initWithZkConfig(config)
}

func (z *ZkDiscovery) initWithZkConfig(c *ZkConfig) error {
	z.running = true
	z.ZkConfig = c
	z.nodes = sync.Map{}
	// todo conf.ini init data to watch
	z.nodes.Range(func(key, value interface{}) bool {
		go z.watchNode(value.(ZkNode))
		return true
	})
	return nil
}

func (z *ZkDiscovery) Watch(key, path string) error {
	if !z.running {
		return errors.New("zk discovery not running")
	}

	if path[0] != '/' {
		path = "/" + path
	}

	zkNode := ZkNode{
		key:      key,
		path:     path,
		stopChan: make(chan struct{}, 1),
	}

	conn, _, err := zk.Connect(z.ZkConfig.host, time.Second*5, zk.WithLogInfo(false))
	if err != nil {
		dlog.Error("watch zk connect occur error:%v", err)
		return err
	}

	zkNode.client = conn

	z.nodes.Store(key, zkNode)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				dlog.Error("etcd Watch panic %s", r)
				return
			}
		}()
		z.watchNode(zkNode)
	}()

	return nil
}

func (z *ZkDiscovery) WatchMulti(nodes map[string]string) error {
	for key, node := range nodes {
		err := z.Watch(key, node)
		if err != nil {
			return err
		}
	}
	return nil
}

func (z *ZkDiscovery) AddNode(key string, info service.NodeInfo) {
	zkNode, ok := z.nodes.Load(key)
	zn := zkNode.(ZkNode)
	if ok {
		nodesInfo := zn.nodesInfo
		nodesInfo = append(nodesInfo, info)
		zn.nodesInfo = nodesInfo
	}
	z.nodes.Store(key, zn)
	return
}

func (z *ZkDiscovery) DelNode(key string, addr string) {
	zkNode, ok := z.nodes.Load(key)
	if !ok {
		return
	}
	zn := zkNode.(ZkNode)
	nodesInfo := zn.nodesInfo
	for k, v := range nodesInfo {
		if v.GetIp()+fmt.Sprintf(":%d", v.GetPort()) == addr {
			nodesInfo = append(nodesInfo[:k], nodesInfo[k+1:]...)
			zn.nodesInfo = nodesInfo
			z.nodes.Store(key, zn)
			break
		}
	}
}

func (z *ZkDiscovery) unMsgNodeInfo(data []byte) service.NodeInfo {
	info := &service.DefaultNodeInfo{}
	err := json.Unmarshal(data, info)
	if err != nil {
		dlog.Error("GetNodeInfo json unmarshal occur derror:%s", err)
		return nil
	}

	return info
}

func (z *ZkDiscovery) GetNodeInfo(key string) []service.NodeInfo {
	nodesInfo, ok := z.nodes.Load(key)
	if !ok {
		return nil
	}
	return nodesInfo.(ZkNode).nodesInfo
}

func (z *ZkDiscovery) watchNode(node ZkNode) {
	nodes, ok := z.nodes.Load(node.key)
	if !ok {
		return
	}

	nodesInfo := nodes.(ZkNode)
	children, _, err := nodesInfo.client.Children(node.path)
	if err != nil {
		dlog.Error("watch node Children occur derror:%s", err)
		return
	}

	for _, v := range children {
		data, _, err := nodesInfo.client.Get(node.path + "/" + v)
		if err != nil {
			dlog.Error("watch node get occur derror:%s", err)
			return
		}
		zkNode := z.unMsgNodeInfo(data)
		z.AddNode(node.key, zkNode)
	}

	for {
		_, _, childCh, err := nodesInfo.client.ChildrenW(node.path)
		if err != nil {
			dlog.Error("watch node watch childrenW occur error:%v", err)
			return
		}

		select {
		case childEvent := <-childCh:
			if childEvent.Type == zk.EventNodeChildrenChanged {
				children, _, err := nodesInfo.client.Children(node.path)
				if err != nil {
					dlog.Error("watch node Children occur error:%v", err)
					return
				}

				z.nodes.Store(node.key, nil)
				for _, v := range children {
					data, _, err := nodesInfo.client.Get(node.path + "/" + v)
					if err != nil {
						dlog.Error("watch node get occur error:%v", err)
						return
					}
					zkNode := z.unMsgNodeInfo(data)
					z.AddNode(node.path, zkNode)
				}
			}
		case <-node.stopChan:
			z.running = false
			return
		}
	}
}
