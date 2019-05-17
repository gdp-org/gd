/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib

import (
	dogError "github.com/chuck1024/godog/error"
	"github.com/chuck1024/godog/utils"
	"github.com/chuck1024/doglog"
	"math/rand"
	"net"
	"sync"
	"time"
)

/*
 * default tcp client
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
		doglog.Error("[AddAddr] parse addr failed, %s", err.Error())
	} else {
		c.addrs = append(c.addrs, addr2)
	}
}

// Stop stop client
func (c *TcpClient) Stop() {
	for addr, cc := range c.Cm {
		cc.Stop()
		doglog.Error("[Stop] stop client %s", addr)
	}

	doglog.Info("[Stop] stop all done.")
}

// connect
func (c *TcpClient) Connect() (*Client, error) {
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
				RequestTimeout: time.Millisecond * time.Duration(c.Timeout),
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
func (c *TcpClient) Invoke(cmd uint32, req []byte, client ...*Client) (rsp []byte, err *dogError.CodeError) {
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
	reqPkt = NewTcpPacket(cmd, req)
	if rspPkt, err = ct.CallRetry(reqPkt, c.RetryNum); err != nil {
		doglog.Error("[Invoke] CallRetry occur error:%v ", err)
		return nil, err
	}

	rsp = rspPkt.(*TcpPacket).Body

	return rsp, nil
}
