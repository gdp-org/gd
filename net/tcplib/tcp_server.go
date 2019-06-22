/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib

import (
	"fmt"
	"github.com/chuck1024/doglog"
	"strconv"
)

/*
 * default tcp server
 */

type Handler func([]byte) (uint32, []byte)

type TcpServer struct {
	addr string
	m    map[uint32]Handler
	ss   *Server
}

func NewTcpServer() *TcpServer {
	s := &TcpServer{
		m: make(map[uint32]Handler),
	}

	s.ss = &Server{
		Handler: s.dispatchPacket,
	}

	return s
}

func (s *TcpServer) Run(port int) error {
	addr := fmt.Sprintf(":%d", port)
	doglog.Info("[Run] Tcp try to listen port: %d", port)

	s.addr = addr
	s.ss.Addr = addr

	err := s.ss.Serve()
	if err != nil {
		doglog.Error("[Run] Start occur error:%s", err.Error())
		return err
	}

	return nil
}

func (s *TcpServer) Stop() {
	s.ss.Stop()
}

func (s *TcpServer) SetAddr(addr string) {
	s.addr = addr
	s.ss.Addr = addr
}

func (s *TcpServer) GetAddr() string {
	return s.addr
}

func (s *TcpServer) AddTcpHandler(headCmd uint32, f Handler) {
	if _, ok := s.m[headCmd]; ok {
		doglog.Warn("[AddTcpHandler] head cmd [%d] already registered.", headCmd)
		return
	}

	s.m[headCmd] = f
	doglog.Info("[AddTcpHandler] register head cmd [%d] success.", headCmd)
}

func (s *TcpServer) dispatchPacket(req Packet) (rsp Packet) {
	packet := req.(*TcpPacket)
	headCmd := packet.Cmd

	f, ok := s.m[headCmd]
	if !ok {
		doglog.Error("[dispatchPacket] head cmd %d not register handler!", headCmd)
		return NewTcpPacketWithRet(headCmd, []byte(""), packet.Seq, uint32(InvalidParam.Code()))
	}

	code, body := GF.Handle(&Context{
		Seq:     packet.Seq,
		Method:  strconv.Itoa(int(headCmd)),
		Handler: f,
		Req:     req.(*TcpPacket).Body,
	})

	return NewTcpPacketWithRet(packet.Cmd, body, packet.Seq, uint32(code))
}
