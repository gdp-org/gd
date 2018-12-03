/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib

import (
	"errors"
	"fmt"
	"github.com/xuyu/logging"
)

/*
 * default tcp server
 */

var (
	AppTcp    *TcpServer
	NoTcpPort = errors.New("no tcp serve port")
)

const NoPort = 0

type Handler func([]byte) (uint32, []byte)

type TcpServer struct {
	addr string
	m    map[uint32]Handler
	ss   *Server
}

func init() {
	AppTcp = NewTcpServer()
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

func (s *TcpServer) Start() error {
	err := s.ss.Serve()
	if err != nil {
		logging.Error("%s", err.Error())
		return err
	}
	return nil
}

func (s *TcpServer) Run(port int) error {
	if port == NoPort {
		logging.Info("[Run] no tcp serve port")
		return NoTcpPort
	}

	addr := fmt.Sprintf(":%d", port)
	logging.Info("[Run] Tcp try to listen port: %d", port)

	s.addr = addr
	s.ss.Addr = addr

	err := s.Start()
	if err != nil {
		logging.Error("[Run] Start occur error:%s", err.Error())
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
		logging.Warning("[AddTcpHandler] head cmd [%d] already registered.", headCmd)
		return
	}

	s.m[headCmd] = f
	logging.Info("[AddTcpHandler] register head cmd [%d] success.", headCmd)
}

func (s *TcpServer) dispatchPacket(req Packet) (rsp Packet) {
	packet := req.(*TcpPacket)
	headCmd := packet.Cmd

	f, ok := s.m[headCmd]
	if !ok {
		logging.Error("[dispatchPacket] head cmd %d not register handler!", headCmd)
		return NewTcpPacketWithRet(headCmd, []byte(""), packet.Seq, uint32(InvalidParam.Code()))
	}

	code, body := f(req.(*TcpPacket).Body)

	return NewTcpPacketWithRet(packet.Cmd, body, packet.Seq, uint32(code))
}
