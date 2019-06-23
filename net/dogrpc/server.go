/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package dogrpc

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

type HandlerFunc func(clientAddr string, req Packet) (rsp Packet)

type Server struct {
	Addr             string
	Handler          HandlerFunc
	Concurrency      int
	FlushDelay       time.Duration
	PendingResponses int
	SendBufferSize   int
	RecvBufferSize   int
	Listener         Listener
	serverStopChan   chan struct{}
	stopWg           sync.WaitGroup
	Encoder          MessageEncoderFunc
	Decoder          MessageDecoderFunc
}

func (s *Server) Start() *dogError.CodeError {
	if s.Handler == nil {
		panic("Server.Handler cannot be nil")
	}

	if s.serverStopChan != nil {
		panic("server is already running. Stop it before starting it again")
	}
	s.serverStopChan = make(chan struct{})

	if s.Concurrency <= 0 {
		s.Concurrency = DefaultConcurrency
	}
	if s.FlushDelay == 0 {
		s.FlushDelay = DefaultFlushDelay
	}
	if s.PendingResponses <= 0 {
		s.PendingResponses = DefaultPendingMessages
	}
	if s.SendBufferSize <= 0 {
		s.SendBufferSize = DefaultBufferSize
	}
	if s.RecvBufferSize <= 0 {
		s.RecvBufferSize = DefaultBufferSize
	}
	if s.Encoder == nil {
		s.Encoder = defaultMessageEncoder
	}
	if s.Decoder == nil {
		s.Decoder = defaultMessageDecoder
	}

	if s.Listener == nil {
		s.Listener = &defaultListener{}
	}

	if err := s.Listener.Init(s.Addr); err != nil {
		ce := InternalServerError.SetMsg(fmt.Sprintf("[%s]. Cannot listen to: [%s]", s.Addr, err))
		return ce
	}

	workersCh := make(chan struct{}, s.Concurrency)
	s.stopWg.Add(1)
	go serverHandler(s, workersCh)
	return nil
}

func (s *Server) Serve() *dogError.CodeError {
	if err := s.Start(); err != nil {
		return err
	}
	s.stopWg.Wait()
	return nil
}

func (s *Server) Stop() {
	if s.serverStopChan == nil {
		panic("server must be started before stopping it")
	}
	close(s.serverStopChan)
	s.stopWg.Wait()
	s.serverStopChan = nil
}

func serverHandler(s *Server, workersCh chan struct{}) {
	defer s.stopWg.Done()

	var conn io.ReadWriteCloser
	var clientAddr string
	var err error
	var stopping atomic.Value

	for {
		acceptChan := make(chan struct{})
		go func() {
			if conn, clientAddr, err = s.Listener.Accept(); err != nil {
				if stopping.Load() == nil {
					doglog.Error("[serverHandler] [%s] cannot accept new connection: [%s]", s.Addr, err)
				}
			}
			close(acceptChan)
		}()

		select {
		case <-s.serverStopChan:
			stopping.Store(true)
			s.Listener.Close()
			<-acceptChan
			return
		case <-acceptChan:
			doglog.Debug("[serverHandler] [%s] connected.", clientAddr)
		}

		if err != nil {
			select {
			case <-s.serverStopChan:
				return
			case <-time.After(time.Second):
				continue
			}
		}

		s.stopWg.Add(1)

		go serverHandleConnection(s, conn, clientAddr, workersCh)
	}
}

func serverHandleConnection(s *Server, conn io.ReadWriteCloser, clientAddr string, workersCh chan struct{}) {
	defer s.stopWg.Done()

	responsesChan := make(chan *serverMessage, s.PendingResponses)
	stopChan := make(chan struct{})
	readerDone := make(chan struct{})
	writerDone := make(chan struct{})

	go serverReader(s, conn, clientAddr, responsesChan, stopChan, readerDone, workersCh)
	go serverWriter(s, conn, clientAddr, responsesChan, stopChan, writerDone)

	select {
	case <-readerDone:
		close(stopChan)
		conn.Close()
		<-writerDone
	case <-writerDone:
		close(stopChan)
		conn.Close()
		<-readerDone
	case <-s.serverStopChan:
		close(stopChan)
		conn.Close()
		<-readerDone
		<-writerDone
	}

	responsesChan = nil
	doglog.Debug("[serverHandleConnection] [%s] disconnected.", clientAddr)
}

func serverReader(s *Server, conn io.ReadWriteCloser, clientAddr string, responsesChan chan<- *serverMessage, stopChan <-chan struct{}, done chan<- struct{}, workersCh chan struct{}) {
	defer func() {
		if r := recover(); r != nil {
			doglog.Error("[serverReader] [%s]->[%s] dumpPanic when reading data from client: %v", clientAddr, s.Addr, r)
		}
		close(done)
	}()

	var err error
	var dec MessageDecoder
	if dec, err = s.Decoder(conn, s.RecvBufferSize); err != nil {
		err = fmt.Errorf("init decoder error:%s", err.Error())
		return
	}

	var req Packet
	for {
		if req, err = dec.Decode(); err != nil {
			if !isClientDisconnect(err) && !isServerStop(stopChan) {
				doglog.Error("[serverReader] [%s] -> [%s] cannot decode request: [%s]", clientAddr, s.Addr, err)
			}
			return
		}

		m := serverMessagePool.Get().(*serverMessage)
		m.Request = req
		m.ClientAddr = clientAddr

		select {
		case workersCh <- struct{}{}:
		default:
			select {
			case workersCh <- struct{}{}:
			case <-stopChan:
				return
			}
		}
		go serverRequest(s, responsesChan, stopChan, m, workersCh)
	}
}

func serverRequest(s *Server, responsesChan chan<- *serverMessage, stopChan <-chan struct{}, m *serverMessage, workersChan <-chan struct{}) {
	req := m.Request
	clientAddr := m.ClientAddr

	m.Request = nil
	m.ClientAddr = ""

	rsp := callHandlerWithRecover(s.Handler, clientAddr, s.Addr, req)
	m.Response = rsp
	select {
	case responsesChan <- m:
	default:
		select {
		case responsesChan <- m:
		case <-stopChan:
		}
	}
	<-workersChan
}

func callHandlerWithRecover(handler HandlerFunc, clientAddr string, serverAddr string, req Packet) (rsp Packet) {
	defer func() {
		if x := recover(); x != nil {
			rsp.SetErrCode(uint32(InternalServerError.Code()))
			stackTrace := make([]byte, 1<<20)
			n := runtime.Stack(stackTrace, false)
			errStr := fmt.Sprintf("Panic occured: %v\n Stack trace: %s", x, stackTrace[:n])
			doglog.Error("[callHandlerWithRecover] [%s] -> [%s]. %s", clientAddr, serverAddr, errStr)
		}
	}()
	rsp = handler(clientAddr, req)
	return
}

func serverWriter(s *Server, conn io.ReadWriteCloser, clientAddr string, responsesChan <-chan *serverMessage, stopChan <-chan struct{}, done chan<- struct{}) {
	defer func() {
		close(done)
	}()

	var err error
	var enc MessageEncoder
	if enc, err = s.Encoder(conn, s.SendBufferSize); err != nil {
		err = fmt.Errorf("init encoder error:%s", err.Error())
		return
	}

	var flushChan <-chan time.Time
	t := time.NewTimer(s.FlushDelay)

	for {
		var m *serverMessage
		select {
		case m = <-responsesChan:
		default:
			runtime.Gosched()
			select {
			case <-stopChan:
				return
			case m = <-responsesChan:
			case <-flushChan:
				if err := enc.Flush(); err != nil {
					if !isServerStop(stopChan) {
						err = fmt.Errorf("[%s] -> [%s] cannot flush response to underlying stream: [%s]", clientAddr, s.Addr, err)
					}
					return
				}
				flushChan = nil
				continue
			}
		}

		if flushChan == nil {
			flushChan = getFlushChan(t, s.FlushDelay)
		}

		rsp := m.Response

		m.Response = nil
		serverMessagePool.Put(m)

		if err := enc.Encode(rsp); err != nil {
			doglog.Error("[serverWriter] [%s] -> [%s] cannot send response: [%s]", clientAddr, s.Addr, err)
			return
		}
	}
}

func isClientDisconnect(err error) bool {
	return err == io.ErrUnexpectedEOF || err == io.EOF
}

func isServerStop(stopChan <-chan struct{}) bool {
	select {
	case <-stopChan:
		return true
	default:
		return false
	}
}

type serverMessage struct {
	Request    Packet
	Response   Packet
	ClientAddr string
}

var serverMessagePool = &sync.Pool{
	New: func() interface{} {
		return &serverMessage{}
	},
}
