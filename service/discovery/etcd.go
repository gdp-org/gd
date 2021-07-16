/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package discovery

// 因为github.com/etcd-io/etcd 最新的能go mod import 的包版本v3.3.25，会导致引入
// github.com/cores/etcd.然后又会引起连锁反应，导致需要使用replace
// github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.4。虽然这种方式也可以使用，但只要使用gd
// 每个项目都会引入replace，极为不优雅。码农网有篇文章写得很好，记录了该问题。https://www.codercto.com/a/108257.html
// etcd作者什么时候把3.4.xx版本能够go mod 能够使用了，再放开etcd的discovery和register
// etcd 3.5.0版本已经修复

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/service"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	"go.etcd.io/etcd/client/v3"
	"gopkg.in/ini.v1"
	"strings"
	"sync"
	"time"
)

type EtcdNode struct {
	key       string // node 的 key，便于获取
	path      string // 节点路径
	nodesInfo []service.NodeInfo
	stopChan  chan struct{}
	client    *clientv3.Client
}

type EtcdConfig struct {
	host      []string // etcd server host
	tlsConfig *tls.Config
}

// Encapsulates the etcd discovery
type EtcdDiscovery struct {
	EtcdConfig   *EtcdConfig `inject:"etcdConfig" canNil:"true"`
	EtcdConf     *ini.File   `inject:"etcdConf" canNil:"true"`
	EtcdConfPath string      `inject:"etcdConfPath" canNil:"true"`

	nodes   sync.Map
	running bool

	startOnce sync.Once
	closeOnce sync.Once
}

func (e *EtcdDiscovery) Start() error {
	var err error
	e.startOnce.Do(func() {
		if e.EtcdConfig != nil {
			err = e.initWithEtcdConfig(e.EtcdConfig)
		} else if e.EtcdConf != nil {
			err = e.initEtcd(e.EtcdConf)
		} else {
			if e.EtcdConfPath == "" {
				e.EtcdConfPath = defaultConf
			}

			err = e.initObjForEtcd(e.EtcdConfPath)
		}
	})
	return err
}

func (e *EtcdDiscovery) Close() {
	e.closeOnce.Do(func() {
		e.nodes.Range(func(key, value interface{}) bool {
			close(value.(EtcdNode).stopChan)
			return true
		})
		return
	})
}

func (e *EtcdDiscovery) initObjForEtcd(filePath string) error {
	etcdConfRealPath := filePath
	if etcdConfRealPath == "" {
		return errors.New("etcdConf not set")
	}

	if !strings.HasSuffix(etcdConfRealPath, ".ini") {
		return errors.New("etcdConf not an ini file")
	}

	etcdConf, err := ini.Load(etcdConfRealPath)
	if err != nil {
		return err
	}

	if err = e.initEtcd(etcdConf); err != nil {
		return err
	}
	return nil
}

func (e *EtcdDiscovery) initEtcd(f *ini.File) error {
	c := f.Section("DisRes")
	hosts := c.Key("etcdHost").Strings(",")

	config := &EtcdConfig{
		host: hosts,
	}

	cert := c.Key("cert").String()
	key := c.Key("key").String()
	ca := c.Key("ca").String()
	if cert != "" && key != "" && ca != "" {
		tlsInfo := transport.TLSInfo{
			CertFile:      cert,
			KeyFile:       key,
			TrustedCAFile: ca,
		}
		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return fmt.Errorf("load tls conf from file fail,, err=%v", err)
		}
		config.tlsConfig = tlsConfig
	}

	return e.initWithEtcdConfig(config)
}

func (e *EtcdDiscovery) initWithEtcdConfig(c *EtcdConfig) error {
	e.nodes = sync.Map{}
	e.EtcdConfig = c
	e.running = true
	// todo conf.ini init data to watch
	e.nodes.Range(func(key, value interface{}) bool {
		go e.watchNode(value.(EtcdNode))
		return true
	})
	return nil
}

func (e *EtcdDiscovery) Watch(key, path string) error {
	if !e.running {
		return errors.New("etcd discovery not running")
	}

	if path[0] != '/' {
		path = "/" + path
	}

	etcdNode := EtcdNode{
		key:      key,
		path:     path,
		stopChan: make(chan struct{}, 1),
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   e.EtcdConfig.host,
		DialTimeout: time.Second,
		TLS:         e.EtcdConfig.tlsConfig,
	})

	if err != nil {
		dlog.Error("watch new client occur error:%s", err)
		return err
	}

	etcdNode.client = cli
	e.nodes.Store(key, etcdNode)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				dlog.Error("etcd Watch panic %s", r)
				return
			}
		}()
		e.watchNode(etcdNode)
	}()
	return nil
}

func (e *EtcdDiscovery) WatchMulti(nodes map[string]string) error {
	for key, node := range nodes {
		err := e.Watch(key, node)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *EtcdDiscovery) AddNode(key string, info service.NodeInfo) {
	etcdNode, ok := e.nodes.Load(key)
	var nodesInfo []service.NodeInfo
	if ok {
		nodesInfo = etcdNode.(EtcdNode).nodesInfo
		nodesInfo = append(nodesInfo, info)
	}
	e.nodes.Store(key, nodesInfo)
	return
}

func (e *EtcdDiscovery) DelNode(key, addr string) {
	etcdNode, ok := e.nodes.Load(key)
	if !ok {
		return
	}
	nodesInfo := etcdNode.(EtcdNode).nodesInfo
	for k, v := range nodesInfo {
		if v.GetIp()+fmt.Sprintf(":%d", v.GetPort()) == addr {
			nodesInfo = append(nodesInfo[:k], nodesInfo[k+1:]...)
			e.nodes.Store(key, nodesInfo)
			break
		}
	}
}

func (e *EtcdDiscovery) unMsgNodeInfo(data []byte) service.NodeInfo {
	info := &service.DefaultNodeInfo{}
	err := json.Unmarshal(data, info)
	if err != nil {
		dlog.Error("GetNodeInfo json unmarshal occur error:%s", err)
		return nil
	}

	return info
}

func (e *EtcdDiscovery) GetNodeInfo(key string) []service.NodeInfo {
	nodesInfo, ok := e.nodes.Load(key)
	if !ok {
		return nil
	}
	return nodesInfo.([]service.NodeInfo)
}

func (e *EtcdDiscovery) watchNode(node EtcdNode) {
	nodes, ok := e.nodes.Load(node.key)
	if !ok {
		return
	}

	nodesInfo := nodes.(EtcdNode)
	resp, err := nodesInfo.client.Get(context.TODO(), node.path, clientv3.WithPrefix())
	if err != nil {
		dlog.Error("watch node get node[%s] children", node.path)
		return
	}

	if resp.Count != 0 {
		for _, ev := range resp.Kvs {
			info := e.unMsgNodeInfo(ev.Value)
			e.AddNode(node.key, info)
		}
	}

	watchChan := nodesInfo.client.Watch(context.Background(), node.path, clientv3.WithPrefix())
	for {
		select {
		case result := <-watchChan:
			for _, ev := range result.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					info := e.unMsgNodeInfo(ev.Kv.Value)
					e.AddNode(node.key, info)
				case clientv3.EventTypeDelete:
					e.DelNode(node.key, string(ev.Kv.Key))
				}
			}
		case <-node.stopChan:
			e.running = false
			return
		}
	}
}
