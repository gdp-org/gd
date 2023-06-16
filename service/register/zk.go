/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package register

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gdp-org/gd/dlog"
	"github.com/gdp-org/gd/service"
	"github.com/gdp-org/gd/utls/network"
	"github.com/samuel/go-zookeeper/zk"
	"gopkg.in/ini.v1"
	"strings"
	"sync"
	"time"
)

type ZkConfig struct {
	Host     []string         // zk server host
	Root     string           // root path
	Group    string           // service group
	Service  string           // service name
	NodeInfo service.NodeInfo // service node info
	Environ  string           //service run environment
}

type ZkRegister struct {
	ZkConfig   *ZkConfig `inject:"zkConfig" canNil:"true"`
	ZkConf     *ini.File `inject:"zkConf" canNil:"true"`
	ZkConfPath string    `inject:"zkConfPath" canNil:"true"`

	client *zk.Conn // zk client

	startOnce sync.Once
	closeOnce sync.Once
}

func (z *ZkRegister) Start() error {
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

func (z *ZkRegister) Close() {
	if z.client != nil {
		z.client.Close()
		z.client = nil
	}
}

func (z *ZkRegister) initObjForZk(filePath string) error {
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

func (z *ZkRegister) initZk(f *ini.File) error {
	c := f.Section("DisRes")
	hosts := c.Key("zkHost").Strings(",")
	root := strings.TrimRight(c.Key("root").String(), "/")
	environ := c.Key("env").String()
	group := c.Key("group").String()

	s := f.Section("Server")
	serviceName := s.Key("serverName").String()

	ip := network.GetLocalIP()
	port := c.Key("regPort").MustInt()
	weight := c.Key("weight").MustUint64()

	config := &ZkConfig{
		Host:    hosts,
		Root:    root,
		Group:   group,
		Service: serviceName,
		Environ: environ,
		NodeInfo: &service.DefaultNodeInfo{
			Ip:      ip,
			Port:    port,
			Offline: false,
			Weight:  weight,
		},
	}

	return z.initWithZkConfig(config)
}

func (z *ZkRegister) initWithZkConfig(c *ZkConfig) error {
	conn, _, err := zk.Connect(c.Host, time.Second*5, zk.WithLogInfo(false))
	if err != nil {
		dlog.Error("zk connect occur error:%s", err)
		return err
	}

	z.ZkConfig = c
	z.client = conn

	err = z.isExistNode()
	if err != nil {
		dlog.Error("zk isExistNode occur error:%s", err)
		return err
	}

	err = z.run()
	if err != nil {
		dlog.Error("zk run occur error:%s", err)
		return err
	}

	return nil
}

func (z *ZkRegister) run() (err error) {
	p := fmt.Sprintf("/%s/%s/%s/%s/pool/%s:%d", z.ZkConfig.Root, z.ZkConfig.Group, z.ZkConfig.Service, z.ZkConfig.Environ,
		z.ZkConfig.NodeInfo.GetIp(), z.ZkConfig.NodeInfo.GetPort())
	dlog.Info("zk path: %s", p)

	dataByte, _ := json.Marshal(&z.ZkConfig.NodeInfo)
	path, err := z.client.Create(p, dataByte, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		dlog.Error("zk create occur error:%s", err)
		return
	}

	if path == p {
		dlog.Info("zk create success! path:%s", path)
	}

	return
}

func (z *ZkRegister) SetOffline(offline bool) {
	z.ZkConfig.NodeInfo.(*service.DefaultNodeInfo).Offline = offline
}

func (z *ZkRegister) SetRootNode(root string) (err error) {
	z.ZkConfig.Root = strings.TrimRight(root, "/")
	if len(z.ZkConfig.Root) == 0 {
		err = fmt.Errorf("invalid root node %s", root)
		return
	}

	return nil
}

func (z *ZkRegister) GetRootNode() (root string) {
	return z.ZkConfig.Root
}

func (z *ZkRegister) SetHeartBeat(heartBeat time.Duration) {
}

func (z *ZkRegister) isExistNode() (err error) {
	node := fmt.Sprintf("/%s/%s/%s", z.ZkConfig.Root, z.ZkConfig.Group, z.ZkConfig.Service)

	isExist, _, err := z.client.Exists(node)
	if err != nil {
		dlog.Error("zk client Exists occur error: %s", err)
		return
	}

	if !isExist {
		p1 := node + "/" + z.ZkConfig.Environ
		p2 := p1 + "/pool"
		paths := []string{node, p1, p2}
		for _, v := range paths {
			path, err := z.client.Create(v, []byte(""), 0, zk.WorldACL(zk.PermAll))
			if err != nil {
				dlog.Error("zk create path occur error: %s, path = %s", err, v)
				return err
			}

			if v != path {
				dlog.Error("zk create path [%s] != path [%s]", node, path)
				return errors.New("rootPath is equal path")
			}
		}
	}

	return
}
