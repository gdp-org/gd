/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib

import (
	"godog/error"
	"time"
)

const (
	DefaultConcurrency     = 8 * 1024
	DefaultRequestTimeout  = 20 * time.Second
	DefaultPendingMessages = 32 * 1024
	DefaultFlushDelay      = -1
	DefaultBufferSize      = 64 * 1024
	DefaultDialRetryTime   = 0
	DefaultConnectNumbers  = 1
)

var (
	TimeOutError        = error.SetCodeType(10001, "timeout error.")
	OverflowError       = error.SetCodeType(10002, "overflow error.")
	InternalServerError = error.SetCodeType(10003, "interval server error.")
	InvalidParam        = error.SetCodeType(10004, "invalid param")
)

var closedFlushChan = make(chan time.Time)

func init() {

	close(closedFlushChan)
}

func getFlushChan(t *time.Timer, flushDelay time.Duration) <-chan time.Time {
	if flushDelay <= 0 {
		return closedFlushChan
	}

	if !t.Stop() {
		// Exhaust expired timer's chan.
		select {
		case <-t.C:
		default:
		}
	}
	t.Reset(flushDelay)
	return t.C
}
