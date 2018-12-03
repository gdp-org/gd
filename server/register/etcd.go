/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package register

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chuck1024/godog/server"
	"github.com/coreos/etcd/clientv3"
	"github.com/xuyu/logging"
	"strings"
	"time"
)

// server path : /root/group/service/environ/pool/ip:port
type EtcdRegister struct {
	host      []string         // etcd server host
	root      string           // root path
	group     string           // service group
	service   string           // service name
	nodeInfo  server.NodeInfo  // service node info
	client    *clientv3.Client // etcd client
	leaseID   clientv3.LeaseID // etcd lease id
	heartBeat uint64           // heartbeat
	exitChan  chan struct{}    // exit signal
	environ   string           // service run environment
}

func (r *EtcdRegister) NewRegister(hosts []string, root, environ, group, service string) {
	r.host = hosts
	r.root = strings.TrimRight(root, "/")
	r.heartBeat = DefaultHeartBeat
	r.exitChan = make(chan struct{})
	r.environ = environ
	r.group = group
	r.service = service

	r.client, _ = clientv3.New(clientv3.Config{
		Endpoints:   r.host,
		DialTimeout: 1 * time.Second,
	})

	return
}

func (r *EtcdRegister) SetOffline(offline bool) {
	r.nodeInfo.(*server.DefaultNodeInfo).Offline = offline
}

func (r *EtcdRegister) SetRootNode(root string) (err error) {
	r.root = strings.TrimRight(root, "/")
	if len(r.root) == 0 {
		err = fmt.Errorf("invalid root node %s", root)
		return
	}

	return nil
}

func (r *EtcdRegister) GetRootNode() (root string) {
	return r.root
}

func (r *EtcdRegister) SetHeartBeat(heartBeat time.Duration) {
	r.heartBeat = uint64(heartBeat)
}

func (r *EtcdRegister) Run(ip string, port int, weight uint64) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logging.Error("[Run] etcd register panic %s", r)
			return
		}
	}()

	ch, err := r.register(ip, port, weight)
	if err != nil {
		logging.Error("[Run] register occur error:%s", err)
		return
	}

	go func() {
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					logging.Debug("[Run] register keep alive channel closed")
					r.revoke()
					return
				}
			case <-r.client.Ctx().Done():
				logging.Warning("[Run] server closed.")
				return
			case <-r.exitChan:
				logging.Debug("[Run] register stop")
				return
			}
		}
	}()
	return
}

func (r *EtcdRegister) register(ip string, port int, weight uint64) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	r.nodeInfo = &server.DefaultNodeInfo{
		Ip:      ip,
		Port:    port,
		Offline: false,
		Weight:  weight,
	}

	node := fmt.Sprintf("%s/%s/%s/%s/pool/%s:%d", r.root, r.group, r.service, r.environ,
		r.nodeInfo.GetIp(), r.nodeInfo.GetPort())

	logging.Info("[register] node:%s", node)

	dataByte, _ := json.Marshal(r.nodeInfo)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	resp, err := r.client.Grant(ctx, int64(r.heartBeat))
	cancel()
	if err != nil {
		logging.Error("[register] client grant occur error:%s", err)
		return nil, err
	}

	for i := 0; i < DefaultRetryTimes; i++ {
		ctx, cancel = context.WithTimeout(context.TODO(), time.Second)
		_, err := r.client.Put(context.TODO(), node, string(dataByte), clientv3.WithLease(resp.ID))
		cancel()
		if err != nil {
			logging.Warning("ectd client set err:%v", err)
			continue
		}

		r.leaseID = resp.ID
		break
	}

	logging.Info("[register] register success!!! service:%s/%s/%s/%s/pool/%s:%d", r.root, r.group, r.service, r.environ,
		r.nodeInfo.GetIp(), r.nodeInfo.GetPort())

	return r.client.KeepAlive(context.TODO(), resp.ID)
}

func (r *EtcdRegister) revoke() error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	_, err := r.client.Revoke(ctx, r.leaseID)
	cancel()
	if err != nil {
		logging.Error("[revoke] occur error:", err)
	}

	logging.Info("[revoke] service:%s/%s/%s/%s/pool/%s:%d", r.root, r.group, r.service, r.environ,
		r.nodeInfo.GetIp(), r.nodeInfo.GetPort())
	return err
}

func (r *EtcdRegister) Close() {
	close(r.exitChan)
	if r.client != nil {
		r.revoke()
		r.client.Close()
		r.client = nil
	}
}
