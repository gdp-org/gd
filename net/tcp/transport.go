/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcp

import (
	"io"
	"net"
	"time"
)

var (
	dialer = &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}
)

type DialFunc func(addr string) (conn io.ReadWriteCloser, err error)

func defaultDial(addr string) (conn io.ReadWriteCloser, err error) {
	return dialer.Dial("tcp", addr)
}
