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

type TcpServer struct {
	Addr string
	m    map[uint32]HandlerFunc
	ss   *Server
}

func init() {
	AppTcpServer = &TcpServer{
		m: make(map[uint32]HandlerFunc),
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
	s.ss.ServerStopChan <- struct{ bool }{true}
}

func (s *TcpServer) AddTcpHandler(headCmd uint32, f HandlerFunc) {
	if _, ok := s.m[headCmd]; ok {
		logging.Warning("[RegisterHandler] head cmd [%d] already registered.", headCmd)
		return
	}

	s.m[headCmd] = f
	logging.Info("[RegisterHandler] register head cmd [%d] success.", headCmd)
}

func (s *TcpServer) dispatchPacket(clientAddr string, req Packet) (rsp Packet) {
	hyPacket := req.(*TcpPacket)
	headCmd := hyPacket.Cmd

	f, ok := s.m[headCmd]
	if !ok {
		logging.Error("[dispatchHyPacket] head cmd %d not register handler!", headCmd)
		return NewCustomPacketWithRet(headCmd, []byte(""), hyPacket.Seq, uint16(InvalidParam.Code()))
	}
	return f(clientAddr, req)
}
