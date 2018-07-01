/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib

import (
	"github.com/xuyu/logging"
)

/*
 * custom server
 */

type CustomServer struct {
	Addr string
	m    map[uint32]HandlerFunc
	ss   *Server
}

func NewServer(addr string) *CustomServer {
	s := &CustomServer{
		Addr: addr,
		m:    make(map[uint32]HandlerFunc),
	}
	s.ss = &Server{
		Addr:    s.Addr,
		Handler: s.dispatchPacket,
	}

	return s
}

func (s *CustomServer) Start() {
	err := s.ss.Start()
	if err != nil {
		logging.Error("%s", err.Error())
	}
}

func (s *CustomServer) Run() {
	err := s.ss.Serve()
	if err != nil {
		logging.Error("%s", err.Error())
	}
}

func (s *CustomServer) Stop() {
	s.ss.Stop()
}

func (s *CustomServer) RegisterTcpHandler(headCmd uint32, f HandlerFunc) {
	if _, ok := s.m[headCmd]; ok {
		logging.Warning("[RegisterHandler] head cmd [%d] already registered.", headCmd)
		return
	}

	s.m[headCmd] = f
	logging.Info("[RegisterHandler] register head cmd [%d] success.", headCmd)
}

func (s *CustomServer) dispatchPacket(clientAddr string, req Packet) (rsp Packet) {
	hyPacket := req.(*CustomPacket)
	headCmd := hyPacket.Cmd

	f, ok := s.m[headCmd]
	if !ok {
		logging.Error("[dispatchHyPacket] head cmd %d not register handler!", headCmd)
		return NewCustomPacketWithRet(headCmd, []byte(""), hyPacket.Seq, uint16(InvalidParam.Code()))
	}
	return f(clientAddr, req)
}
