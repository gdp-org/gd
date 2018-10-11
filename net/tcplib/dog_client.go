/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib

import (
	"bufio"
	dogError "github.com/chuck1024/godog/error"
	"github.com/xuyu/logging"
	"io"
	"math/rand"
	"net"
	"time"
)

/*
 * dog client
 */

// dog packet. Invoke rpc call
func (c *TcpClient) DogInvoke(cmd uint32, req []byte) (rsp []byte, err *dogError.CodeError) {
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
			logging.Warning("[Invoke] Addr %s already created.", addr)
		}
	} else {
		if cc.clientStopChan == nil {
			cc.Start()
		}
	}

	var reqPkt, rspPkt Packet
	reqPkt = NewDogPacket(cmd, req)
	if rspPkt, err = cc.CallRetry(reqPkt, c.RetryNum); err != nil {
		logging.Error("[Invoke] CallRetry occur error:%v ", err)
		return nil, err
	}

	rsp = rspPkt.(*DogPacket).Body

	return rsp, nil
}
