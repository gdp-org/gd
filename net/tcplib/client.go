/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package tcplib

import (
	"fmt"
	"github.com/chuck1024/doglog"
	dogError "github.com/chuck1024/godog/error"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Client struct {
	Addr                 string
	Conns                int
	Dial                 DialFunc
	DialRetryTime        int
	PendingRequests      int
	FlushDelay           time.Duration
	RequestTimeout       time.Duration
	SendBufferSize       int
	RecvBufferSize       int
	pendingRequestsCount uint32
	requestsChan         chan *AsyncResult
	clientStopChan       chan struct{}
	stopWg               sync.WaitGroup
	stopLock             sync.Mutex
	startLock            sync.Mutex
	Encoder              MessageEncoderFunc
	Decoder              MessageDecoderFunc
}

func (c *Client) Start() {
	doglog.Info("start %s", c.Addr)
	defer c.startLock.Unlock()
	c.startLock.Lock()
	if c.clientStopChan != nil {
		doglog.Warn("[Start]: the given client is already started. Call Client.Stop() before calling Client.Start() again!")
	}

	if c.Conns <= 0 {
		c.Conns = DefaultConnectNumbers
	}
	if c.Dial == nil {
		c.Dial = defaultDial
	}
	if c.DialRetryTime < 0 {
		c.DialRetryTime = DefaultDialRetryTime
	}
	if c.PendingRequests <= 0 {
		c.PendingRequests = DefaultPendingMessages
	}
	if c.FlushDelay == 0 {
		c.FlushDelay = DefaultFlushDelay
	}
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = DefaultRequestTimeout
	}
	if c.SendBufferSize <= 0 {
		c.SendBufferSize = DefaultBufferSize
	}
	if c.RecvBufferSize <= 0 {
		c.RecvBufferSize = DefaultBufferSize
	}
	if c.DialRetryTime < 0 {
		c.DialRetryTime = DefaultDialRetryTime
	}

	c.requestsChan = make(chan *AsyncResult, c.PendingRequests)
	c.clientStopChan = make(chan struct{})

	if c.Encoder == nil {
		c.Encoder = defaultMessageEncoder
	}
	if c.Decoder == nil {
		c.Decoder = defaultMessageDecoder
	}

	for i := 0; i < c.Conns; i++ {
		c.stopWg.Add(1)
		go clientHandler(c)
	}
}

func (c *Client) Stop() {
	defer c.stopLock.Unlock()
	c.stopLock.Lock()
	if c.clientStopChan == nil {
		doglog.Error("[Stop]: the client must be started before stopping it")
	}
	close(c.clientStopChan)
	c.stopWg.Wait()
	c.clientStopChan = nil
}

func clientHandler(c *Client) {
	defer func() {
		c.stopWg.Done()
		c.Stop()
	}()

	var conn io.ReadWriteCloser
	var err error
	var stopping atomic.Value
	var dialRetryTime = c.DialRetryTime

	for {
		dialChan := make(chan struct{})
		// connect
		go func() {
			if conn, err = c.Dial(c.Addr); err != nil {
				if stopping.Load() == nil {
					doglog.Error("[clientHandler]>> cannot establish connection to [%s], error [%s]", c.Addr, err)
				}
			}
			close(dialChan)
		}()

		// wait connecting... if client was stopped, notice the establishment of a connection firstly, and then return when it connected
		select {
		case <-c.clientStopChan:
			stopping.Store(true)
			<-dialChan // could remove?
			return
		case <-dialChan:
		}

		// connect fail
		if err != nil {
			// need to reconnect
			if dialRetryTime > 0 {
				select { // if client is already stop, quit
				case <-c.clientStopChan:
					return
				case <-time.After(time.Second):
					dialRetryTime--
					continue
				}
			} else {
				return
			}
		}

		clientHandleConnection(c, conn)

		dialRetryTime = c.DialRetryTime

		// It happens abnormal when connection doesn't closed. need to closed connection.
		select {
		case <-c.clientStopChan:
			return
		default:
		}
	}
}

func clientHandleConnection(c *Client, conn io.ReadWriteCloser) {
	stopChan := make(chan struct{})
	writerDone := make(chan error, 1)
	readerDone := make(chan error, 1)

	pendingRequests := make(map[uint32]*AsyncResult)
	var pendingRequestLock sync.Mutex

	go clientWriter(c, conn, pendingRequests, &pendingRequestLock, stopChan, writerDone)
	go clientReader(c, conn, pendingRequests, &pendingRequestLock, readerDone)

	var err error
	select {
	case err = <-writerDone:
		close(stopChan)
		conn.Close()
		<-readerDone
	case err = <-readerDone:
		close(stopChan)
		conn.Close()
		<-writerDone
	case <-c.clientStopChan:
		close(stopChan)
		conn.Close()
		<-readerDone
		<-writerDone
	}

	if err != nil {
		doglog.Error("[clientHandleConnection] occur error: ", c.Addr+", "+err.Error())
	}

	for _, m := range pendingRequests {
		atomic.AddUint32(&c.pendingRequestsCount, ^uint32(0))
		m.Error = err
		if m.Done != nil {
			close(m.Done)
		}
	}
}

func clientWriter(c *Client, conn io.Writer, pendingRequests map[uint32]*AsyncResult, pendingRequestLock *sync.Mutex, stopChan <-chan struct{}, done chan<- error) {
	var err error
	defer func() {
		done <- err
	}()

	var enc MessageEncoder
	if enc, err = c.Encoder(conn, c.SendBufferSize); err != nil {
		err = fmt.Errorf("Init encoder error: %s ", err.Error())
		return
	}

	t := time.NewTimer(c.FlushDelay)
	var flushChan <-chan time.Time
	for {
		var m *AsyncResult
		select {
		case m = <-c.requestsChan:
		default:
			runtime.Gosched()

			select {
			case <-stopChan:
				return
			case m = <-c.requestsChan:
			case <-flushChan:
				if err = enc.Flush(); err != nil {
					err = fmt.Errorf("Cannot flush requests to underlying stream:%s [%s] ", c.Addr, err)
					return
				}
				flushChan = nil
				continue
			}
		}

		if flushChan == nil {
			flushChan = getFlushChan(t, c.FlushDelay)
		}

		// check connect is canceled.
		if m.isCanceled() {
			if m.Done != nil {
				m.Error = fmt.Errorf("canceled error")
				close(m.Done)
			} else {
				releaseAsyncResult(m)
			}
			continue
		}

		if !m.isSkipResponse() {
			msgID := m.Request.ID()

			pendingRequestLock.Lock()
			n := len(pendingRequests)
			pendingRequests[msgID] = m
			pendingRequestLock.Unlock()
			atomic.AddUint32(&c.pendingRequestsCount, 1)

			if n > 3*c.PendingRequests {
				// timeout Clear connect
				doglog.Info("pending requests(%d), clean canceled req...", n)
				pendingRequestLock.Lock()
				for _, m := range pendingRequests {
					atomic.AddUint32(&c.pendingRequestsCount, ^uint32(0))
					if m.isCanceled() {
						delete(pendingRequests, msgID)
					}
				}
				pendingRequestLock.Unlock()
				n = len(pendingRequests)
				if n > 3*c.PendingRequests {
					err = fmt.Errorf(" The server [%s] didn't return %d responses yet. Closing server connection in order to prevent client resource leaks", c.Addr, n)
					return
				}
			}
		}

		if err = enc.Encode(m.Request); err != nil {
			return
		}
	}
}

func clientReader(c *Client, conn io.Reader, pendingRequests map[uint32]*AsyncResult, pendingRequestLock *sync.Mutex, done chan<- error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			if err == nil {
				err = fmt.Errorf(" Panic when reading data from server[%s]: %v", c.Addr, r)
			}
		}
		done <- err
	}()

	var dec MessageDecoder
	if dec, err = c.Decoder(conn, c.RecvBufferSize); err != nil {
		err = fmt.Errorf(" Init decoder error:%s", err.Error())
		return
	}
	for {
		var packet Packet
		if packet, err = dec.Decode(); err != nil {
			err = fmt.Errorf("decode packet error: %s", err.Error())
			return
		}

		msgID := packet.ID()
		pendingRequestLock.Lock()
		m, ok := pendingRequests[msgID]
		if ok {
			delete(pendingRequests, msgID)
		}
		pendingRequestLock.Unlock()
		if !ok {
			err = fmt.Errorf("unexpected msgID=[%d] obtained from server [%s]", msgID, c.Addr)
			continue
		}

		atomic.AddUint32(&c.pendingRequestsCount, ^uint32(0))
		m.Response = packet

		close(m.Done)
	}
}

func (c *Client) Call(req Packet) (rsp Packet, err *dogError.CodeError) {
	return c.CallTimeout(req, c.RequestTimeout, 0)
}

func (c *Client) CallRetry(req Packet, retryNum uint32) (rsp Packet, err *dogError.CodeError) {
	return c.CallTimeout(req, c.RequestTimeout, retryNum)
}

// skip response
func (c *Client) SendUDP(req Packet) (err *dogError.CodeError) {
	if _, err = c.CallAsync(req, true); err != nil {
		return err
	}
	return nil
}

func (c *Client) CallTimeout(req Packet, timeout time.Duration, retryNum uint32) (rsp Packet, err *dogError.CodeError) {
	var tryNum uint32
retry:
	var m *AsyncResult
	if m, err = c.callAsync(req, false, true); err != nil {
		return nil, err
	}
	t := acquireTimer(timeout)
	select {
	case <-m.Done:
		if m.Error == nil {
			rsp, err = m.Response, nil
		} else {
			rsp, err = m.Response, InternalServerError.SetMsg(m.Error.Error())
		}
		releaseAsyncResult(m)
	case <-t.C:
		m.Cancel()
		err = TimeOutError.SetMsg(fmt.Sprintf("[%s]. Cannot obtain response during timeout=%s", c.Addr, timeout))
	}
	releaseTimer(t)

	if err != nil && err.Code() == TimeOutError.Code() {
		tryNum++
		if tryNum <= retryNum {
			goto retry
		}
	}

	return
}

func (c *Client) CallAsync(req Packet, skipResponse bool) (*AsyncResult, *dogError.CodeError) {
	return c.callAsync(req, skipResponse, false)
}

func (c *Client) callAsync(req Packet, skipResponse bool, usePool bool) (m *AsyncResult, err *dogError.CodeError) {
	if skipResponse {
		usePool = true
	}
	if usePool {
		m = acquireAsyncResult()
	} else {
		m = &AsyncResult{}
	}

	m.Request = req
	m.skipResponse = skipResponse
	if !skipResponse {
		m.t = time.Now()
		m.Done = make(chan struct{})
	}

	select {
	case c.requestsChan <- m:
		return m, nil
	default:
		if skipResponse {
			releaseAsyncResult(m)
			return nil, TimeOutError
		}
		select {
		case mm := <-c.requestsChan:
			if mm.Done != nil {
				mm.Error = OverflowError
				close(mm.Done)
			} else {
				if usePool {
					releaseAsyncResult(mm)
				}
			}
		default:
		}
		select {
		case c.requestsChan <- m:
			return m, nil
		default:
			if usePool {
				releaseAsyncResult(m)
			}
			return nil, OverflowError
		}
	}
}

type AsyncResult struct {
	Response     Packet
	Error        error
	Done         chan struct{}
	Request      Packet
	t            time.Time
	canceled     uint32
	skipResponse bool
}

func (m *AsyncResult) Cancel() {
	atomic.StoreUint32(&m.canceled, 1)
}

func (m *AsyncResult) isCanceled() bool {
	return atomic.LoadUint32(&m.canceled) != 0
}

func (m *AsyncResult) isSkipResponse() bool {
	return m.skipResponse
}

var asyncResultPool sync.Pool

func acquireAsyncResult() *AsyncResult {
	v := asyncResultPool.Get()
	if v == nil {
		return &AsyncResult{}
	}
	return v.(*AsyncResult)
}

var zeroTime time.Time

func releaseAsyncResult(m *AsyncResult) {
	m.Response = nil
	m.Error = nil
	m.Done = nil
	m.Request = nil
	m.t = zeroTime
	asyncResultPool.Put(m)
}

var timerPool sync.Pool

func acquireTimer(timeout time.Duration) *time.Timer {
	tv := timerPool.Get()
	if tv == nil {
		return time.NewTimer(timeout)
	}

	t := tv.(*time.Timer)
	if t.Reset(timeout) {
		doglog.Error("[acquireTimer] BUG: Active timer trapped into acquireTimer()")
	}
	return t
}

func releaseTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}

	timerPool.Put(t)
}
