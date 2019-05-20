/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package httplib

import (
	"errors"
	"fmt"
	"github.com/chuck1024/doglog"
	"net/http"
)

type InitHandlerFunc func() error
type HandlerFunc func(http.ResponseWriter, *http.Request)

var (
	AppHttp    *HttpServer
	NoHttpPort = errors.New("no http serve port")
)

const NoPort = 0

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

func init() {
	AppHttp = NewHttpServer()
}

type HttpServer struct {
	health      Handler
	handler     Handler
	initHandler InitHandlerFunc
	handlerMap  map[string]HandlerFunc
}

func NewHttpServer() *HttpServer {
	return &HttpServer{
		health:      nil,
		handler:     nil,
		initHandler: nil,
		handlerMap:  make(map[string]HandlerFunc),
	}
}

func (h *HttpServer) SetHealthHandler(handler Handler) {
	h.health = handler
}

func (h *HttpServer) SetServeHandler(handler Handler) {
	h.handler = handler
}

func (h *HttpServer) SetInitHandler(handler InitHandlerFunc) {
	h.initHandler = handler
}

func (h *HttpServer) JudgeInitHandler() bool {
	if h.initHandler == nil {
		return false
	}
	return true
}

func (h *HttpServer) Serve(httpPort int, handler http.Handler) {
	srvPort := fmt.Sprintf(":%d", httpPort)
	doglog.Info("[Serve] Http try to listen port: %d", httpPort)
	go func() {
		err := http.ListenAndServe(srvPort, handler)
		if err != nil {
			doglog.Error("[Serve] Listen failed, error = %s", err.Error())
			return
		}
	}()
}

func (h *HttpServer) Health(healthPort int, handler http.Handler) {
	srvPort := fmt.Sprintf("%d", healthPort)
	doglog.Info("[Health] Try to monitor health condition on port: %s", srvPort)
	go func() {
		err := http.ListenAndServe(srvPort, handler)
		if err != nil {
			doglog.Error("[Health] monitor failed, error = %s", err.Error())
			return
		}
	}()
}

func (h *HttpServer) HandleFunc(addr string, handler HandlerFunc) {
	http.HandleFunc(addr, handler)
}

func (h *HttpServer) AddHttpHandler(addr string, handler HandlerFunc) {
	_, ok := h.handlerMap[addr]
	if ok {
		doglog.Warn("[AddHandlerFunc] Try to replace handler to addr = %s", addr)
	}

	h.handlerMap[addr] = handler
	doglog.Info("[AddHandlerFunc] Add/Replace [addr: %s] ok", addr)
}

func (h *HttpServer) Register() {
	for k, v := range h.handlerMap {
		h.HandleFunc(k, v)
	}
}

func (h *HttpServer) Run(healthPort, port int) error {
	// http health
	if healthPort != NoPort && h.health != nil {
		h.Health(healthPort, h.health)
	}

	// http service
	if port == NoPort {
		doglog.Info("[Run] No http Serve port for application ")
		return NoHttpPort
	} else {
		h.Serve(port, h.handler)
	}

	return nil
}
