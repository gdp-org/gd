/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package httplib

import (
	"errors"
	"fmt"
	"github.com/chuck1024/godog/config"
	"github.com/xuyu/logging"
	"net/http"
)

type InitHandlerFunc func() error
type HandlerFunc func(http.ResponseWriter, *http.Request)

var (
	AppHttp    *HttpServer
	NoHttpPort = errors.New("no http serve port")
	NoPort     = 0
)

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

func init(){
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
	logging.Info("[Serve] Http try to listen port: %d", httpPort)
	go func() {
		err := http.ListenAndServe(srvPort, handler)
		if err != nil {
			logging.Error("[Serve] Listen failed, error = %s", err.Error())
			return
		}
	}()
}

func (h *HttpServer) Health(healthPort int, handler http.Handler) {
	srvPort := fmt.Sprintf("%d", healthPort)
	logging.Info("[Health] Try to monitor health condition on port: %s", srvPort)
	go func() {
		err := http.ListenAndServe(srvPort, handler)
		if err != nil {
			logging.Error("[Health] monitor failed, error = %s", err.Error())
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
		logging.Warning("[AddHandlerFunc] Try to replace handler to addr = %s", addr)
	}

	h.handlerMap[addr] = handler
	logging.Info("[AddHandlerFunc] Add/Replace [addr: %s] ok", addr)
}

func (h *HttpServer) Register() {
	for k, v := range h.handlerMap {
		h.HandleFunc(k, v)
	}
}

func (h *HttpServer) Run() error {
	// http health
	if config.AppConfig.BaseConfig.Prog.HealthPort != NoPort && h.health != nil {
		h.Health(config.AppConfig.BaseConfig.Prog.HealthPort, h.health)
	}

	// http service
	if config.AppConfig.BaseConfig.Server.HttpPort == NoPort {
		logging.Info("[Run] No http Serve port for application ")
		return NoHttpPort
	} else {
		h.Serve(config.AppConfig.BaseConfig.Server.HttpPort, h.handler)
	}

	return nil
}
