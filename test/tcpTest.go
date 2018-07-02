/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/xuyu/logging"
	"godog/net/tcplib"
	"time"
)

func server() {
	s := tcplib.AppTcpServer

	s.AddTcpHandler(1024, func(clientAddr string, req tcplib.Packet) (rsp tcplib.Packet) {
		cReq := req.(*tcplib.TcpPacket)
		rsp = tcplib.NewCustomPacketWithSeq(cReq.Cmd, []byte("1024 hello."), cReq.Seq)
		return
	})

	go func() {
		time.Sleep(10 * time.Second)
		s.Stop()
	}()

	s.Run()
}

func client() {
	c := tcplib.NewClient(500, 0)
	c.AddAddr("192.168.1.107:10241")

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
