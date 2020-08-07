/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc

import (
	"github.com/chuck1024/dlog"
	dogError "github.com/chuck1024/gd/derror"
	"github.com/chuck1024/gd/utls/network"
	"math/rand"
	"net"
	"sync"
	"time"
)

/*
 * default rpc client
 */

type RpcClient struct {
	Cm       map[string]*Client
	cmMutex  sync.Mutex
	addrs    []*net.TCPAddr
	Timeout  time.Duration
	RetryNum uint32
	localIp  string
}

func NewClient(timeout time.Duration, retryNum uint32) *RpcClient {
	return &RpcClient{
		Cm:       make(map[string]*Client),
		Timeout:  timeout,
		RetryNum: retryNum,
		localIp:  network.GetLocalIP(),
	}
}

// add server address
func (c *RpcClient) AddAddr(addr string) {
	if addr2, err := net.ResolveTCPAddr("tcp", addr); err != nil {
		dlog.Error("parse addr failed, %s", err.Error())
	} else {
		c.addrs = append(c.addrs, addr2)
	}
}

// Stop stop client
func (c *RpcClient) Stop() {
	for addr, cc := range c.Cm {
		cc.Stop()
		dlog.Error("dog rpc client stop client %s", addr)
	}

	dlog.Info("dog rpc client stop all done.")
}

// connect
func (c *RpcClient) Connect() (*Client, error) {
	addr := &net.TCPAddr{}

	if len(c.addrs) > 0 {
		rand.Seed(time.Now().UnixNano())
		idx := rand.Intn(len(c.addrs))
		addr = c.addrs[idx]
	} else {
		return nil, InternalServerError
	}

	cc, ok := c.Cm[addr.String()]
	if !ok {
		c.cmMutex.Lock()
		defer c.cmMutex.Unlock()
		if cc, ok = c.Cm[addr.String()]; !ok {
			cc = &Client{
				Addr:           addr.String(),
				RequestTimeout: c.Timeout,
			}
			cc.Start()
			c.Cm[addr.String()] = cc
		} else {
			dlog.Warn("[Connect] Addr %s already created.", addr)
		}
	} else {
		if cc.clientStopChan == nil {
			cc.Start()
		}
	}

	return cc, nil
}

// Invoke rpc call
func (c *RpcClient) Invoke(cmd uint32, req []byte, client ...*Client) (code uint32, rsp []byte, err *dogError.CodeError) {
	var ct *Client
	if len(client) == 0 {
		cc, err := c.Connect()
		if err != nil {
			dlog.Error("[Invoke] connect occur error:%s", err)
			return code, nil, InternalServerError
		}
		ct = cc
	} else {
		ct = client[0]
	}

	var reqPkt, rspPkt Packet
	reqPkt = NewRpcPacket(cmd, req)
	if rspPkt, err = ct.CallRetry(reqPkt, c.RetryNum); err != nil {
		dlog.Error("[Invoke] CallRetry occur error:%v ", err)
		return code, nil, err
	}

	rsp = rspPkt.(*RpcPacket).Body
	code = rspPkt.(*RpcPacket).ErrCode

	return code, rsp, nil
}
