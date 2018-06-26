/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcp

import (
	"fmt"
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

type CodeError struct {
	code int
	msg  string
}

func (e *CodeError) Code() int {
	return e.code
}

func (e *CodeError) Error() string {
	return fmt.Sprintf("[%d] %s", e.code, e.msg)
}

func (e *CodeError) Msg(msg string) *CodeError {
	e.msg = msg
	return e
}

func NewCodeError(code int, msg string) *CodeError {
	return &CodeError{
		code: code,
		msg:  msg,
	}
}

var (
	TimeOutError        = &CodeError{10001, "timeout error."}
	OverflowError       = &CodeError{10002, "overflow error."}
	InternalServerError = &CodeError{10003, "interval server error."}
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
