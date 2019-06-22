/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc

import (
	"github.com/chuck1024/doglog"
	dogError "github.com/chuck1024/godog/error"
	"github.com/chuck1024/godog/utils"
	"math/rand"
	"net"
	"sync"
	"time"
)

/*
 * default tcp client
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
		localIp:  utils.GetLocalIP(),
	}
}

// add server address
func (c *RpcClient) AddAddr(addr string) {
	if addr2, err := net.ResolveTCPAddr("tcp", addr); err != nil {
		doglog.Error("[AddAddr] parse addr failed, %s", err.Error())
	} else {
		c.addrs = append(c.addrs, addr2)
	}
}

// Stop stop client
func (c *RpcClient) Stop() {
	for addr, cc := range c.Cm {
		cc.Stop()
		doglog.Error("[Stop] stop client %s", addr)
	}

	doglog.Info("[Stop] stop all done.")
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
			doglog.Warn("[Connect] Addr %s already created.", addr)
		}
	} else {
		if cc.clientStopChan == nil {
			cc.Start()
		}
	}

	return cc, nil
}

// Invoke rpc call
func (c *RpcClient) Invoke(cmd uint32, req []byte, client ...*Client) (rsp []byte, err *dogError.CodeError) {
	var ct *Client
	if len(client) == 0 {
		cc, err := c.Connect()
		if err != nil {
			doglog.Error("[Invoke] connect occur error:%s", err)
			return nil, InternalServerError
		}
		ct = cc
	} else {
		ct = client[0]
	}

	var reqPkt, rspPkt Packet
	reqPkt = NewRpcPacket(cmd, req)
	if rspPkt, err = ct.CallRetry(reqPkt, c.RetryNum); err != nil {
		doglog.Error("[Invoke] CallRetry occur error:%v ", err)
		return nil, err
	}

	rsp = rspPkt.(*RpcPacket).Body

	return rsp, nil
}
