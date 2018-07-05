/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"sync/atomic"
)

type Packet interface {
	ID() uint32
	SetErrCode(code uint32)
}

type MessageEncoder interface {
	Encode(msg Packet) error
	Flush() error
}

type MessageDecoder interface {
	Decode() (Packet, error)
}

type MessageEncoderFunc func(w io.Writer, bufferSize int) (encoder MessageEncoder, err error)
type MessageDecoderFunc func(r io.Reader, bufferSize int) (decoder MessageDecoder, err error)

func defaultMessageEncoder(w io.Writer, bufferSize int) (encoder MessageEncoder, err error) {
	return &TcpPacketEncoder{bw: bufio.NewWriterSize(w, bufferSize)}, nil
}

func defaultMessageDecoder(r io.Reader, bufferSize int) (decoder MessageDecoder, err error) {
	return &TcpPacketDecoder{br: bufio.NewReaderSize(r, bufferSize)}, nil
}

/*
 * TcpPacket.
 */

const (
	HeaderLen = 13
	SohLen    = 1
	EohLen    = 2
)

type TcpPacket struct {
	SOH uint8
	Header
	Body []byte
	EOH  uint8
}

type Header struct {
	Seq       uint32
	ErrCode   uint16
	Cmd       uint32
	PacketLen uint32
	Version   uint8
	CheckSum  uint16
}

var (
	globalSeq uint32
)

const (
	SOH = 0x06
	EOH = 0x07
)

func nextSeq() uint32 {
	return atomic.AddUint32(&globalSeq, 1)
}

func (p *TcpPacket) ID() uint32 {
	return p.Seq
}

func (p *TcpPacket) SetErrCode(code uint32) {
	p.ErrCode = uint16(code)
}

func NewTcpPacket(cmd uint32, body []byte) *TcpPacket {
	seq := nextSeq()
	return NewTcpPacketWithSeq(cmd, body, seq)
}

func NewTcpPacketWithSeq(cmd uint32, body []byte, seq uint32) *TcpPacket {
	return NewTcpPacketWithRet(cmd, body, seq, 0)
}

func NewTcpPacketWithRet(cmd uint32, body []byte, seq uint32, ret uint16) *TcpPacket {
	return &TcpPacket{
		SOH: SOH,
		Header: Header{
			Version:   0,
			Cmd:       cmd,
			CheckSum:  0,
			Seq:       seq,
			ErrCode:   ret,
			PacketLen: uint32(len(body)) + HeaderLen + 2},
		Body: body,
		EOH:  EOH,
	}
}

type TcpPacketEncoder struct {
	bw *bufio.Writer
}

type TcpPacketDecoder struct {
	br *bufio.Reader
}

func (e *TcpPacketEncoder) Encode(p Packet) error {
	if packet, ok := p.(*TcpPacket); ok {
		if err := binary.Write(e.bw, binary.BigEndian, packet.SOH); err != nil {
			return err
		}
		if err := binary.Write(e.bw, binary.BigEndian, packet.Header); err != nil {
			return err
		}
		if err := binary.Write(e.bw, binary.BigEndian, packet.Body); err != nil {
			return err
		}
		if err := binary.Write(e.bw, binary.BigEndian, packet.EOH); err != nil {
			return err
		}

		return nil
	}
	return errors.New("SelfPacketEncoder Encode occur error")
}

func (e *TcpPacketEncoder) Flush() error {
	if e.bw != nil {
		if err := e.bw.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// of course, Decode Function need you to judge packet SOH, EOH and packet length.
func (d *TcpPacketDecoder) Decode() (Packet, error) {
	packet := &TcpPacket{}

	if err := binary.Read(d.br, binary.BigEndian, &packet.SOH); err != nil {
		return nil, err
	}

	if err := binary.Read(d.br, binary.BigEndian, &packet.Header); err != nil {
		return nil, err
	}

	bodyLen := packet.PacketLen - HeaderLen - SohLen - EohLen
	packet.Body = make([]byte, bodyLen)
	if err := binary.Read(d.br, binary.BigEndian, packet.Body); err != nil {
		return nil, err
	}

	if err := binary.Read(d.br, binary.BigEndian, &packet.EOH); err != nil {
		return nil, err
	}

	return packet, nil
}
