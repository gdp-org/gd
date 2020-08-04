/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc

import (
	"bufio"
	"github.com/chuck1024/dlog"
	dogError "github.com/chuck1024/godog/error"
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
func (c *RpcClient) DogInvoke(cmd uint32, req []byte, client ...*Client) (code uint32, rsp []byte, err *dogError.CodeError) {
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

	var reqPkt, rspPkt Packet
	reqPkt = NewDogPacket(cmd, req)
	if rspPkt, err = ct.CallRetry(reqPkt, c.RetryNum); err != nil {
		dlog.Error("Invoke CallRetry occur error:%v ", err)
		return code, nil, err
	}

	rsp = rspPkt.(*DogPacket).Body
	code = rspPkt.(*DogPacket).ErrCode

	return code, rsp, nil
}
