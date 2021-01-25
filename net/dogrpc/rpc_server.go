/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc

import (
	"fmt"
	"github.com/chuck1024/gd/dlog"
	"strconv"
)

/*
 * default rpc server
 */

type RpcHandlerFunc func([]byte) (uint32, []byte)

type RpcServer struct {
	Addr           int `inject:"rpcHost"`
	ss             *Server
	defaultHandler map[uint32]RpcHandlerFunc
	wrapHandler    map[uint32]interface{}
}

func NewRpcServer() *RpcServer {
	s := &RpcServer{
		defaultHandler: make(map[uint32]RpcHandlerFunc),
	}

	s.ss = &Server{
		Handler: s.dispatchPacket,
	}

	return s
}

func (s *RpcServer) Start() error {
	s.ss.Addr = fmt.Sprintf(":%d", s.Addr)

	err := s.ss.Serve()
	if err != nil {
		dlog.Error("dog Start occur error:%s", err.Error())
		return err
	}

	return nil
}

func (s *RpcServer) Close() {
	s.ss.Stop()
}

func (s *RpcServer) AddHandler(headCmd uint32, f RpcHandlerFunc) {
	if s.defaultHandler == nil {
		s.defaultHandler = make(map[uint32]RpcHandlerFunc)
	}

	if _, ok := s.defaultHandler[headCmd]; ok {
		dlog.Warn("add handler head cmd [%d] already registered.", headCmd)
		return
	}

	s.defaultHandler[headCmd] = f
	dlog.Info("register head cmd [%d] success.", headCmd)
}

func (s *RpcServer) dispatchPacket(clientAddr string, req Packet) (rsp Packet) {
	packet := req.(*RpcPacket)
	headCmd := packet.Cmd

	f, ok := s.defaultHandler[headCmd]
	if !ok {
		dlog.Error("dispatch packet head cmd %d not register handler!", headCmd)
		return NewRpcPacketWithRet(headCmd, []byte(""), packet.Seq, uint32(InvalidParam.Code()))
	}

	code, body := globalFilter.Handle(&Context{
		ClientAddr: clientAddr,
		Seq:        packet.Seq,
		Method:     strconv.Itoa(int(headCmd)),
		Handler:    f,
		Req:        req.(*RpcPacket).Body,
	})

	return NewRpcPacketWithRet(packet.Cmd, body, packet.Seq, uint32(code))
}
