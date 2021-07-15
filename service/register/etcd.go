/**
 * Copyright 2018 gd Authoe. All Rights Reserved.
 * Author: Chuck1024
 */

package register

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
	"github.com/chuck1024/gd/runtime/inject"
	"github.com/chuck1024/gd/service"
	"github.com/chuck1024/gd/utls"
	"github.com/chuck1024/gd/utls/network"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	"go.etcd.io/etcd/client/v3"
	"gopkg.in/ini.v1"
	"strings"
	"sync"
	"time"
)

// server path : /root/group/service/environ/pool/ip:port
type EtcdConfig struct {
	host      []string         // etcd server host
	root      string           // root path
	group     string           // service group
	service   string           // service name
	nodeInfo  service.NodeInfo // service node info
	heartBeat uint64           // heartbeat
	environ   string           // service run environment
	tlsConfig *tls.Config
}

type EtcdRegister struct {
	EtcdConfig   *EtcdConfig `inject:"etcdConfig" canNil:"true"`
	EtcdConf     *ini.File   `inject:"etcdConf" canNil:"true"`
	EtcdConfPath string      `inject:"etcdConfPath" canNil:"true"`

	client   *clientv3.Client // etcd client
	leaseID  clientv3.LeaseID // etcd lease id
	exitChan chan struct{}    // exit signal

	startOnce sync.Once
	closeOnce sync.Once
}

func (e *EtcdRegister) Start() error {
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

func (e *EtcdRegister) Close() {
	e.closeOnce.Do(func() {
		close(e.exitChan)
		if e.client != nil {
			e.revoke()
			e.client.Close()
			e.client = nil
		}
	})
}

func (e *EtcdRegister) initObjForEtcd(filePath string) error {
	etcdConfRealPath := filePath
	if etcdConfRealPath == "" {
		return errors.New("etcdConf not set in g_cfg")
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

func (e *EtcdRegister) initEtcd(f *ini.File) error {
	c := f.Section("DisRes")
	hosts := c.Key("etcdHost").Strings(",")
	root := strings.TrimRight(c.Key("root").String(), "/")

	heartBeat := c.Key("heartBeat").MustUint64()
	if heartBeat == 0 {
		heartBeat = DefaultHeartBeat
	}

	environ := c.Key("environ").String()
	group := c.Key("group").String()

	s := f.Section("Server")
	serviceName := s.Key("serverName").String()

	ip := network.GetLocalIP()
	port, ok := inject.Find("regPort")
	if !ok {
		if s.Key("httpPort").MustInt() > 0 {
			port = s.Key("httpPort").MustInt()
		} else if s.Key("rpcPort").MustInt() > 0 {
			port = s.Key("rpcPort").MustInt()
		} else if s.Key("grpcPort").MustInt() > 0 {
			port = s.Key("grpcPort").MustInt()
		}
	}
	weight := c.Key("weight").MustUint64()

	config := &EtcdConfig{
		host:      hosts,
		root:      root,
		group:     group,
		service:   serviceName,
		heartBeat: heartBeat,
		environ:   environ,
		nodeInfo: &service.DefaultNodeInfo{
			Ip:      ip,
			Port:    int(utls.MustInt64(port, 0)),
			Offline: false,
			Weight:  weight,
		},
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

func (e *EtcdRegister) initWithEtcdConfig(c *EtcdConfig) error {
	e.EtcdConfig = c
	e.exitChan = make(chan struct{})
	e.client, _ = clientv3.New(clientv3.Config{
		Endpoints:   c.host,
		DialTimeout: 1 * time.Second,
		TLS:         c.tlsConfig,
	})

	ch, err := e.register()
	if err != nil {
		dlog.Error("etcd register occur derror:%s", err)
		return err
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				dlog.Error("etcd register panic %s", r)
				return
			}
		}()

		for {
			select {
			case _, ok := <-ch:
				if !ok {
					dlog.Debug("etcd register keep alive channel closed")
					e.revoke()
					return
				}
			case <-e.client.Ctx().Done():
				dlog.Warn("etcd server closed.")
				return
			case <-e.exitChan:
				dlog.Debug("etcd register stop")
				return
			}
		}
	}()
	return nil
}

func (e *EtcdRegister) register() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	node := fmt.Sprintf("%s/%s/%s/%s/pool/%s:%d", e.EtcdConfig.root, e.EtcdConfig.group, e.EtcdConfig.service, e.EtcdConfig.environ,
		e.EtcdConfig.nodeInfo.GetIp(), e.EtcdConfig.nodeInfo.GetPort())

	dlog.Info("etcd register node:%s", node)

	dataByte, _ := json.Marshal(e.EtcdConfig.nodeInfo)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	resp, err := e.client.Grant(ctx, int64(e.EtcdConfig.heartBeat))
	cancel()
	if err != nil {
		dlog.Error("etcd register client grant occur derror:%s", err)
		return nil, err
	}

	for i := 0; i < DefaultRetryTimes; i++ {
		ctx, cancel = context.WithTimeout(context.TODO(), time.Second)
		_, err := e.client.Put(context.TODO(), node, string(dataByte), clientv3.WithLease(resp.ID))
		cancel()
		if err != nil {
			dlog.Warn("ectd client set err:%v", err)
			continue
		}

		e.leaseID = resp.ID
		break
	}

	dlog.Info("register success!!! service:%s/%s/%s/%s/pool/%s:%d", e.EtcdConfig.root, e.EtcdConfig.group, e.EtcdConfig.service, e.EtcdConfig.environ,
		e.EtcdConfig.nodeInfo.GetIp(), e.EtcdConfig.nodeInfo.GetPort())

	return e.client.KeepAlive(context.TODO(), resp.ID)
}

func (e *EtcdRegister) revoke() error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	_, err := e.client.Revoke(ctx, e.leaseID)
	cancel()
	if err != nil {
		dlog.Error("revoke occur derror:", err)
	}

	dlog.Info("revoke service:%s/%s/%s/%s/pool/%s:%d", e.EtcdConfig.root, e.EtcdConfig.group, e.EtcdConfig.service, e.EtcdConfig.environ,
		e.EtcdConfig.nodeInfo.GetIp(), e.EtcdConfig.nodeInfo.GetPort())
	return err
}

func (e *EtcdRegister) SetOffline(offline bool) {
	e.EtcdConfig.nodeInfo.(*service.DefaultNodeInfo).Offline = offline
}

func (e *EtcdRegister) SetRootNode(root string) (err error) {
	e.EtcdConfig.root = strings.TrimRight(root, "/")
	if len(e.EtcdConfig.root) == 0 {
		err = fmt.Errorf("invalid root node %s", root)
		return
	}

	return nil
}

func (e *EtcdRegister) GetRootNode() (root string) {
	return e.EtcdConfig.root
}

func (e *EtcdRegister) SetHeartBeat(heartBeat time.Duration) {
	e.EtcdConfig.heartBeat = uint64(heartBeat)
}
