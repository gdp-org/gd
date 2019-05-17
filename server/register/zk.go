/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package register

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chuck1024/doglog"
	"github.com/chuck1024/godog/server"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"time"
)

type ZkRegister struct {
	host     []string        // zk server host
	root     string          // root path
	group    string          // service group
	service  string          // service name
	nodeInfo server.NodeInfo // service node info
	client   *zk.Conn        // zk client
	environ  string          //service run environment
}

func (z *ZkRegister) NewRegister(hosts []string, root, environ, group, service string) {
	z.host = hosts
	z.root = strings.TrimRight(root, "/")
	z.environ = environ
	z.group = group
	z.service = service

	conn, _, err := zk.Connect(hosts, time.Second*5, zk.WithLogInfo(false))
	if err != nil {
		doglog.Error("[NewRegister] zk connect occur error:%s", err)
		return
	}

	z.client = conn
	doglog.Debug("[NewRegister] connect success")
	return
}

func (z *ZkRegister) SetOffline(offline bool) {
	z.nodeInfo.(*server.DefaultNodeInfo).Offline = offline
}

func (z *ZkRegister) SetRootNode(root string) (err error) {
	z.root = strings.TrimRight(root, "/")
	if len(z.root) == 0 {
		err = fmt.Errorf("invalid root node %s", root)
		return
	}

	return nil
}

func (z *ZkRegister) GetRootNode() (root string) {
	return z.root
}

func (z *ZkRegister) SetHeartBeat(heartBeat time.Duration) {
}

func (z *ZkRegister) isExistNode() (err error) {
	node := fmt.Sprintf("%s/%s/%s", z.root, z.group, z.service)

	isExist, _, err := z.client.Exists(node)
	if err != nil {
		doglog.Error("[isExistNode] client Exists occur error: %s", err)
		return
	}

	if !isExist {
		p1 := node + "/" + z.environ
		p2 := p1 + "/pool"
		paths := []string{node, p1, p2}
		for _, v := range paths {
			path, err := z.client.Create(v, []byte(""), 0, zk.WorldACL(zk.PermAll))
			if err != nil {
				doglog.Error("[isExistNode] create path occur error: %s", err)
				return err
			}

			if v != path {
				doglog.Error("[isExistNode] create path [%s] != path [%s]", node, path)
				return errors.New("rootPath is equal path")
			}
		}
	}

	return
}

func (z *ZkRegister) Run(ip string, port int, weight uint64) (err error) {
	defer func() {
		if r := recover(); r != nil {
			doglog.Error("[Run] zk register panic %s", r)
			return
		}
	}()

	z.nodeInfo = &server.DefaultNodeInfo{
		Ip:      ip,
		Port:    port,
		Offline: false,
		Weight:  weight,
	}

	err = z.isExistNode()
	if err != nil {
		doglog.Error("[Run] isExistNode occur error:%s", err)
		return
	}

	err = z.run()
	if err != nil {
		doglog.Error("[Run] run occur error:%s", err)
		return
	}

	return
}

func (z *ZkRegister) run() (err error) {
	p := fmt.Sprintf("%s/%s/%s/%s/pool/%s:%d", z.root, z.group, z.service, z.environ,
		z.nodeInfo.GetIp(), z.nodeInfo.GetPort())
	doglog.Info("[run] p: %s", p)

	dataByte, _ := json.Marshal(&z.nodeInfo)
	path, err := z.client.Create(p, dataByte, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		doglog.Error("[run] create occur error:%s", err)
		return
	}

	if path == p {
		doglog.Info("[run] create success! path:%s", path)
	}

	return
}

func (z *ZkRegister) Close() {
	if z.client != nil {
		z.client.Close()
		z.client = nil
	}
}
