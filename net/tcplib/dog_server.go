/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib

import (
	"bufio"
	"github.com/chuck1024/doglog"
	"io"
)

/*
 * dog server
 */

var (
	AppDog *TcpServer
)

func init() {
	AppDog = NewDogTcpServer()
}

func NewDogTcpServer() *TcpServer {
	s := &TcpServer{
		m: make(map[uint32]Handler),
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

func (s *TcpServer) dogDispatchPacket(req Packet) (rsp Packet) {
	packet := req.(*DogPacket)
	headCmd := packet.Cmd

	f, ok := s.m[headCmd]
	if !ok {
		doglog.Error("[dispatchPacket] head cmd %d not register handler!", headCmd)
		return NewDogPacketWithRet(headCmd, []byte(""), packet.Seq, uint32(InvalidParam.Code()))
	}

	code, body := f(req.(*DogPacket).Body)

	return NewDogPacketWithRet(packet.Cmd, body, packet.Seq, uint32(code))
}
