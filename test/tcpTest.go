/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/xuyu/logging"
	"godog/net/tcplib"
)

func server() {
	s := tcplib.NewServer("127.0.0.1:1024")
	defer s.Stop()
	s.RegisterTcpHandler(1024, func(clientAddr string, req tcplib.Packet) (rsp tcplib.Packet) {
		cReq := req.(*tcplib.CustomPacket)
		rsp = tcplib.NewCustomPacketWithSeq(cReq.Cmd, []byte("1024 hello."), cReq.Seq)
		return
	})

	s.Start()
}

func client() {
	c := &tcplib.Client{
		Addr: "127.0.0.1:1024",
	}

	var reqPkt, rspPkt tcplib.Packet
	body := []byte("test success")
	reqPkt = tcplib.NewCustomPacket(1024, body)

	c.Start()

	rspPkt, err := c.Call(reqPkt)
	if err != nil {
		logging.Error("Error when sending request to server: %s", err)
	}
	rsp := rspPkt.(*tcplib.CustomPacket).Body
	logging.Debug("resp=%s", string(rsp))
}

func main() {
	server()
}
