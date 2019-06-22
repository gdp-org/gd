/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc

import (
	"fmt"
	"github.com/chuck1024/doglog"
	"strconv"
)

/*
 * default rpc server
 */

type Handler func([]byte) (uint32, []byte)

type RpcServer struct {
	addr string
	m    map[uint32]Handler
	ss   *Server
}

func NewRpcServer() *RpcServer {
	s := &RpcServer{
		m: make(map[uint32]Handler),
	}

	s.ss = &Server{
		Handler: s.dispatchPacket,
	}

	return s
}

func (s *RpcServer) Run(port int) error {
	addr := fmt.Sprintf(":%d", port)
	doglog.Info("[Run] rpc try to listen port: %d", port)

	s.addr = addr
	s.ss.Addr = addr

	err := s.ss.Serve()
	if err != nil {
		doglog.Error("[Run] Start occur error:%s", err.Error())
		return err
	}

	return nil
}

func (s *RpcServer) Stop() {
	s.ss.Stop()
}

func (s *RpcServer) SetAddr(addr string) {
	s.addr = addr
	s.ss.Addr = addr
}

func (s *RpcServer) GetAddr() string {
	return s.addr
}

func (s *RpcServer) AddHandler(headCmd uint32, f Handler) {
	if _, ok := s.m[headCmd]; ok {
		doglog.Warn("[AddHandler] head cmd [%d] already registered.", headCmd)
		return
	}

	s.m[headCmd] = f
	doglog.Info("[AddHandler] register head cmd [%d] success.", headCmd)
}

func (s *RpcServer) dispatchPacket(req Packet) (rsp Packet) {
	packet := req.(*RpcPacket)
	headCmd := packet.Cmd

	f, ok := s.m[headCmd]
	if !ok {
		doglog.Error("[dispatchPacket] head cmd %d not register handler!", headCmd)
		return NewRpcPacketWithRet(headCmd, []byte(""), packet.Seq, uint32(InvalidParam.Code()))
	}

	code, body := GF.Handle(&Context{
		Seq:     packet.Seq,
		Method:  strconv.Itoa(int(headCmd)),
		Handler: f,
		Req:     req.(*RpcPacket).Body,
	})

	return NewRpcPacketWithRet(packet.Cmd, body, packet.Seq, uint32(code))
}