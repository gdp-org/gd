/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"godog/net/tcplib"
	"github.com/xuyu/logging"
	"sync/atomic"
)

var globalSeq uint32
func nextSeq() uint32 {
	return atomic.AddUint32(&globalSeq, 1)
}

func server(){
	s := &tcplib.Server{
		Addr:"127.0.0.1:16666",
		Handler: func(clientAddr string, req tcplib.Packet)tcplib.Packet{
			logging.Debug("Obtained request %+v from the client %s\n", req, clientAddr)
			body := []byte("niu bi")
			rsp := NewPacket(body)
			return rsp
		},
	}

	if err := s.Serve(); err != nil {
		logging.Error("Cannot stop godog tcp server:%s",err)
	}
}

func client(){
	c := &tcplib.Client{
		Addr: "127.0.0.1:16666",
	}

	var reqPkt, rspPkt tcplib.Packet
	body := []byte("test success")
	reqPkt = NewPacket(body)

	c.Start()

	rspPkt, err := c.Call(reqPkt)
	if err != nil {
		logging.Error("Error when sending request to server: %s", err)
	}
	rsp := rspPkt.(*tcplib.CustomPacket).Body
	logging.Debug("resp=%s",string(rsp))
}

func NewPacket(body []byte) *tcplib.CustomPacket {
	seq := nextSeq()
	return &tcplib.CustomPacket{
		SOH: 6,
		Header: tcplib.Header{
			Version:      0,
			CheckSum:     0,
			MsgID:        seq,
			ErrCode:      0,
			PacketLen:    uint32(len(body)) + tcplib.HeaderLen + 2},
		Body: body,
		EOH:  8,
	}
}

func main(){
	client()
}