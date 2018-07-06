/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib

import (
	"github.com/xuyu/logging"
	"godog/utils"
	"math/rand"
	"net"
	"sync"
	"time"
)

/*
 * tcp client
 */

var (
	AppTcpClient *TcpClient
)

type TcpClient struct {
	Cm       map[string]*Client
	cmMutex  sync.Mutex
	addrs    []*net.TCPAddr
	Timeout  uint32
	RetryNum uint32
	localIp  string
}

func NewClient(timeout, retryNum uint32) *TcpClient {
	AppTcpClient = &TcpClient{
		Cm:       make(map[string]*Client),
		Timeout:  timeout,
		RetryNum: retryNum,
		localIp:  utils.GetLocalIP(),
	}

	return AppTcpClient
}

// add server address
func (c *TcpClient) AddAddr(addr string) {
	if addr2, err := net.ResolveTCPAddr("tcp", addr); err != nil {
		logging.Error("[AddAddr] parse addr failed, %s", err.Error())
	} else {
		c.addrs = append(c.addrs, addr2)
	}
}

// Stop stop client
func (c *TcpClient) Stop() {
	for addr, cc := range c.Cm {
		cc.Stop()
		logging.Error("[Stop] stop client %s", addr)
	}

	logging.Info("[Stop] stop all done.")
}

// Invoke rpc call
func (c *TcpClient) Invoke(cmd uint32, req []byte) (rsp []byte, err *CodeError) {
	addr := &net.TCPAddr{}

	if len(c.addrs) > 0 {
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
				RequestTimeout: time.Millisecond * time.Duration(c.Timeout),
			}
			cc.Start()
			c.Cm[addr.String()] = cc
		} else {
			logging.Warning("[Invoke] Addr %s already created.", addr)
		}
	} else {
		if cc.clientStopChan == nil {
			cc.Start()
		}
	}

	var reqPkt, rspPkt Packet
	reqPkt = NewTcpPacket(cmd, req)
	if rspPkt, err = cc.CallRetry(reqPkt, c.RetryNum); err != nil {
		logging.Error("[Invoke] CallRetry occur error:%v ", err)
		return nil, err
	}

	rsp = rspPkt.(*TcpPacket).Body

	return rsp, nil
}
