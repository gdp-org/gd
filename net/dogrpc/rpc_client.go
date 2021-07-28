/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc

import (
	"crypto/tls"
	"crypto/x509"
	dogError "github.com/chuck1024/gd/derror"
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/utls/network"
	"io"
	"io/ioutil"
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

	TlsCfg           *tls.Config
	RpcCaPemFile     string
	RpcClientKeyFile string
	RpcClientPemFile string
}

func DefaultNewClient(timeout time.Duration, retryNum uint32, useTls bool) *RpcClient {
	r := &RpcClient{
		Cm:       make(map[string]*Client),
		Timeout:  timeout,
		RetryNum: retryNum,
		localIp:  network.GetLocalIP(),
	}

	if useTls && r.TlsCfg == nil {
		if r.RpcCaPemFile == "" {
			r.RpcCaPemFile = "conf/rpc-ca.pem"
		}

		if r.RpcClientKeyFile == "" {
			r.RpcClientKeyFile = "conf/rpc-client.key"
		}

		if r.RpcClientPemFile == "" {
			r.RpcClientPemFile = "conf/rpc-client.pem"
		}

		cert, err := tls.LoadX509KeyPair(r.RpcClientPemFile, r.RpcClientKeyFile)
		if err != nil {
			return nil
		}

		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(r.RpcCaPemFile)
		if err != nil {
			return nil
		}

		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return nil
		}

		r.TlsCfg = &tls.Config{
			RootCAs:            certPool,
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		}
	}

	return r
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

			if c.TlsCfg != nil {
				cc.Dial = func(addr string) (conn io.ReadWriteCloser, err error) {
					c, err := tls.DialWithDialer(dialer, "tcp", addr, c.TlsCfg)
					if err != nil {
						return nil, err
					}
					return c, err
				}
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
