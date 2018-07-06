/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib

import (
	"errors"
	"fmt"
	"github.com/xuyu/logging"
	"godog/config"
	"godog/utils"
)

/*
 * tcp server
 */

var (
	AppTcpServer *TcpServer
	NoTcpPort    = errors.New("no tcp serve port")
)

type Handler func([]byte) (uint16, []byte)

type TcpServer struct {
	Addr string
	m    map[uint32]Handler
	ss   *Server
}

func init() {
	AppTcpServer = &TcpServer{
		m: make(map[uint32]Handler),
	}
	AppTcpServer.ss = &Server{
		Handler: AppTcpServer.dispatchPacket,
	}
}

func (s *TcpServer) Start() {
	err := s.ss.Start()
	if err != nil {
		logging.Error("%s", err.Error())
	}
}

func (s *TcpServer) Run() error {
	port := config.AppConfig.BaseConfig.Server.TcpPort
	if port == 0 {
		logging.Info("[Run] no tcp serve port")
		return NoTcpPort
	}

	localIp := utils.GetLocalIP()
	addr := fmt.Sprintf("%s:%d", localIp, port)
	logging.Info("[tcpServer] Try to listen: %s", addr)

	s.Addr = addr
	s.ss.Addr = addr

	err := s.ss.Serve()
	if err != nil {
		logging.Error("%s", err.Error())
		return err
	}

	return nil
}

func (s *TcpServer) Stop() {
	s.ss.serverStopChan <- struct{ bool }{true}
}

func (s *TcpServer) AddTcpHandler(headCmd uint32, f Handler) {
	if _, ok := s.m[headCmd]; ok {
		logging.Warning("[RegisterHandler] head cmd [%d] already registered.", headCmd)
		return
	}

	s.m[headCmd] = f
	logging.Info("[RegisterHandler] register head cmd [%d] success.", headCmd)
}

func (s *TcpServer) dispatchPacket(req Packet) (rsp Packet) {
	packet := req.(*TcpPacket)
	headCmd := packet.Cmd

	f, ok := s.m[headCmd]
	if !ok {
		logging.Error("[dispatchPacket] head cmd %d not register handler!", headCmd)
		return NewTcpPacketWithRet(headCmd, []byte(""), packet.Seq, uint16(InvalidParam.Code()))
	}

	code, body := f(req.(*TcpPacket).Body)

	return NewTcpPacketWithRet(packet.Cmd, body, packet.Seq, code)
}
