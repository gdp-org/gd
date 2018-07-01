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
	s := tcplib.NewServer("127.0.0.1:10240")
	defer s.Stop()
	s.RegisterTcpHandler(1024, func(clientAddr string, req tcplib.Packet) (rsp tcplib.Packet) {
		cReq := req.(*tcplib.CustomPacket)
		rsp = tcplib.NewCustomPacketWithSeq(cReq.Cmd, []byte("1024 hello."), cReq.Seq)
		return
	})

	s.Run()
}

func client() {
	c := tcplib.NewClient(500, 0)
	c.AddAddr("127.0.0.1:10240")

	body := []byte("test success")

	rsp, err := c.Invoke(1024, body)
	if err != nil {
		logging.Error("Error when sending request to server: %s", err)
	}

	logging.Debug("resp=%s", string(rsp))
}

func main() {
	client()
}
