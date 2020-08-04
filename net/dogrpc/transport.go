/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc

import (
	"io"
	"net"
	"time"
)

const (
	DefaultDialNetWork = "tcp"
)

var (
	dialer = &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}
)

type DialFunc func(addr string) (conn io.ReadWriteCloser, err error)

type Listener interface {
	Init(addr string) error
	Accept() (conn io.ReadWriteCloser, clientAddr string, err error)
	Close() error
	ListenAddr() net.Addr
}

func defaultDial(addr string) (conn io.ReadWriteCloser, err error) {
	return dialer.Dial(DefaultDialNetWork, addr)
}

type defaultListener struct {
	L net.Listener
}

func (ln *defaultListener) Init(addr string) (err error) {
	ln.L, err = net.Listen(DefaultDialNetWork, addr)
	return
}

func (ln *defaultListener) ListenAddr() net.Addr {
	if ln.L != nil {
		return ln.L.Addr()
	}
	return nil
}

func (ln *defaultListener) Accept() (conn io.ReadWriteCloser, clientAddr string, err error) {
	c, err := ln.L.Accept()
	if err != nil {
		return nil, "", err
	}
	if err = setupKeepalive(c); err != nil {
		c.Close()
		return nil, "", err
	}
	return c, c.RemoteAddr().String(), nil
}

func (ln *defaultListener) Close() error {
	return ln.L.Close()
}

func setupKeepalive(conn net.Conn) error {
	tcpConn := conn.(*net.TCPConn)
	if err := tcpConn.SetKeepAlive(true); err != nil {
		return err
	}
	if err := tcpConn.SetKeepAlivePeriod(30 * time.Second); err != nil {
		return err
	}
	return nil
}
