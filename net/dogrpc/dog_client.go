/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	dogError "github.com/gdp-org/gd/derror"
	"github.com/gdp-org/gd/dlog"
	"io"
	"math/rand"
	"net"
	"time"
)

/*
 * dog client
 */

// dog packet establish connection
func (c *RpcClient) DogConnect() (*Client, error) {
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
				Encoder: func(w io.Writer, bufferSize int) (encoder MessageEncoder, err error) {
					return &DogPacketEncoder{bw: bufio.NewWriterSize(w, bufferSize)}, nil
				},
				Decoder: func(r io.Reader, bufferSize int) (decoder MessageDecoder, err error) {
					return &DogPacketDecoder{br: bufio.NewReaderSize(r, bufferSize)}, nil
				},
			}

			if c.TlsCfg != nil {
				cc.Dial = func(addr string) (conn io.ReadWriteCloser, err error) {
					c, err := tls.DialWithDialer(dialer, DefaultDialNetWork, addr, c.TlsCfg)
					if err != nil {
						return nil, err
					}
					return c, err
				}
			}

			cc.Start()
			c.Cm[addr.String()] = cc
		} else {
			dlog.Warn("Addr %s already created.", addr)
		}
	} else {
		if cc.clientStopChan == nil {
			cc.Start()
		}
	}

	return cc, nil
}

// dog packet. Invoke rpc call
func (c *RpcClient) DogInvoke(cmd uint32, req interface{}, client ...*Client) (code uint32, rsp []byte, err *dogError.CodeError) {
	var ct *Client
	if len(client) == 0 {
		cc, err := c.DogConnect()
		if err != nil {
			dlog.Error("Invoke connect occur error:%s", err)
			return code, nil, InternalServerError
		}
		ct = cc
	} else {
		ct = client[0]
	}

	var body []byte
	if req != nil {
		body, _ = json.Marshal(req)
	}

	var reqPkt, rspPkt Packet
	reqPkt = NewDogPacket(cmd, body)
	if rspPkt, err = ct.CallRetry(reqPkt, c.RetryNum); err != nil {
		dlog.Error("Invoke CallRetry occur error:%v ", err)
		return code, nil, err
	}

	rsp = rspPkt.(*DogPacket).Body
	code = rspPkt.(*DogPacket).ErrCode

	return code, rsp, nil
}
