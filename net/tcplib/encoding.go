/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash/crc32"
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
	HeaderLen = 20
	Version   = 1
	Padding   = 0
)

type TcpPacket struct {
	Header
	Body []byte
}

type Header struct {
	PacketLen uint32
	Seq       uint32
	Cmd       uint32
	CheckSum  uint32
	ErrCode   uint16
	Version   uint8
	Padding   uint8
	SOH       uint8
	EOH       uint8
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
	packet := &TcpPacket{
		Header: Header{
			PacketLen: uint32(len(body)) + HeaderLen,
			Seq:       seq,
			Cmd:       cmd,
			CheckSum:  0,
			ErrCode:   ret,
			Version:   Version,
			Padding:   Padding,
			SOH:       SOH,
			EOH:       EOH,
		},
		Body: body,
	}

	packetByte, _ := json.Marshal(packet)
	checkSum := crc32.ChecksumIEEE(packetByte)
	packet.CheckSum = checkSum

	return packet
}

type TcpPacketEncoder struct {
	bw *bufio.Writer
}

type TcpPacketDecoder struct {
	br *bufio.Reader
}

func (e *TcpPacketEncoder) Encode(p Packet) error {
	if packet, ok := p.(*TcpPacket); ok {
		if err := binary.Write(e.bw, binary.BigEndian, packet.Header); err != nil {
			return err
		}
		if err := binary.Write(e.bw, binary.BigEndian, packet.Body); err != nil {
			return err
		}

		return nil
	}
	return errors.New("TcpPacketEncoder Encode occur error")
}

func (d *TcpPacketDecoder) Decode() (Packet, error) {
	packet := &TcpPacket{}

	if err := binary.Read(d.br, binary.BigEndian, &packet.Header); err != nil {
		return nil, err
	}

	if packet.Header.PacketLen < HeaderLen {
		return nil, errors.New("invalid packet")
	}

	if packet.Header.SOH != SOH {
		return nil, errors.New("invalid SOH")
	}

	if packet.Header.EOH != EOH {
		return nil, errors.New("invalid EOH")
	}

	bodyLen := packet.Header.PacketLen - HeaderLen
	packet.Body = make([]byte, bodyLen)

	if err := binary.Read(d.br, binary.BigEndian, packet.Body); err != nil {
		return nil, err
	}

	checkSum1 := packet.Header.CheckSum
	packet.Header.CheckSum = 0
	packetByte, _ := json.Marshal(packet)
	checkSum2 := crc32.ChecksumIEEE(packetByte)

	if checkSum1 != checkSum2 {
		return nil, errors.New("invalid CheckSum")
	}

	packet.Header.CheckSum = checkSum1

	return packet, nil
}

func (e *TcpPacketEncoder) Flush() error {
	if e.bw != nil {
		if err := e.bw.Flush(); err != nil {
			return err
		}
	}
	return nil
}
