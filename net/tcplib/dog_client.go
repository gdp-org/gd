/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib

import (
	"bufio"
	"github.com/chuck1024/doglog"
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
func (c *TcpClient) DogConnect() (*Client, error) {
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
			doglog.Warn("[Connect] Addr %s already created.", addr)
		}
	} else {
		if cc.clientStopChan == nil {
			cc.Start()
		}
	}

	return cc, nil
}

// dog packet. Invoke rpc call
func (c *TcpClient) DogInvoke(cmd uint32, req []byte, client ...*Client) (rsp []byte, err *dogError.CodeError) {
	var ct *Client
	if len(client) == 0 {
		cc, err := c.DogConnect()
		if err != nil {
			doglog.Error("[DogInvoke] connect occur error:%s", err)
			return nil, InternalServerError
		}
		ct = cc
	} else {
		ct = client[0]
	}

	var reqPkt, rspPkt Packet
	reqPkt = NewDogPacket(cmd, req)
	if rspPkt, err = ct.CallRetry(reqPkt, c.RetryNum); err != nil {
		doglog.Error("[Invoke] CallRetry occur error:%v ", err)
		return nil, err
	}

	rsp = rspPkt.(*DogPacket).Body

	return rsp, nil
}
