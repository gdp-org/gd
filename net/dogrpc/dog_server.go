/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc

import (
	"bufio"
	"github.com/chuck1024/gd/dlog"
	"io"
	"strconv"
)

/*
 * dog server
 */

func NewDogRpcServer() *RpcServer {
	s := &RpcServer{
		defaultHandler: make(map[uint32]RpcHandlerFunc),
	}

	s.ss = &Server{
		Handler: s.dogDispatchPacket,
		Encoder: func(w io.Writer, bufferSize int) (encoder MessageEncoder, err error) {
			return &DogPacketEncoder{bw: bufio.NewWriterSize(w, bufferSize)}, nil
		},
		Decoder: func(r io.Reader, bufferSize int) (decoder MessageDecoder, err error) {
			return &DogPacketDecoder{br: bufio.NewReaderSize(r, bufferSize)}, nil
		},
	}

	return s
}

func (s *RpcServer) AddDogHandler(headCmd uint32, f interface{}) {
	if s.wrapHandler == nil {
		s.wrapHandler = make(map[uint32]interface{})
	}

	if _, ok := s.wrapHandler[headCmd]; ok {
		dlog.Warn("[AddDogHandler] wrapHandler head cmd [%d] already registered.", headCmd)
		return
	}

	s.wrapHandler[headCmd] = f
	dlog.Info("AddDogHandler wrapHandler register head cmd [%d] success.", headCmd)
}

func (s *RpcServer) DogRpcRegister() error {
	for k, v := range s.wrapHandler {
		wf, err := wrap(v)
		if err != nil {
			dlog.Error("DogRpcRegister wrap occur error:%s", err)
			return err
		}
		s.AddHandler(k, wf)
	}
	return nil
}

func (s *RpcServer) dogDispatchPacket(clientAddr string, req Packet) (rsp Packet) {
	packet := req.(*DogPacket)
	headCmd := packet.Cmd

	f, ok := s.defaultHandler[headCmd]
	if !ok {
		dlog.Error("dispatchPacket head cmd %d not register handler!", headCmd)
		return NewDogPacketWithRet(headCmd, []byte(""), packet.Seq, uint32(InvalidParam.Code()))
	}

	code, body := globalFilter.Handle(&Context{
		ClientAddr: clientAddr,
		Seq:        packet.Seq,
		Method:     strconv.Itoa(int(headCmd)),
		Handler:    f,
		Req:        req.(*DogPacket).Body,
	})

	return NewDogPacketWithRet(packet.Cmd, body, packet.Seq, code)
}
