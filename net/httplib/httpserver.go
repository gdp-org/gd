/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package httplib

import (
	"errors"
	"github.com/xuyu/logging"
	"godog/config"
)

type InitHandlerFunc func() error

var (
	AppHttp    *HttpServer
	NoHttpPort = errors.New("no http serve port")
)

type HttpServer struct {
	health      Handler
	handler     Handler
	initHandler InitHandlerFunc
	handlerMap  map[string]HandlerFunc
}

func init() {
	AppHttp = &HttpServer{
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

func (h *HttpServer) AddHandlerFunc(addr string, handler HandlerFunc) {
	_, ok := h.handlerMap[addr]
	if ok {
		logging.Warning("[AddHandlerFunc] Try to replace handler to addr = %s", addr)
	}

	h.handlerMap[addr] = handler
	logging.Info("[AddHandlerFunc] Add/Replace [addr: %s] ok", addr)
}

func (h *HttpServer) Register() {
	for k, v := range h.handlerMap {
		HandleFunc(k, v)
	}
}

func (h *HttpServer) Run() error {
	// http health
	if config.AppConfig.BaseConfig.Prog.HealthPort != 0 && h.health != nil {
		Health(config.AppConfig.BaseConfig.Prog.HealthPort, h.health)
	}

	// http service
	if config.AppConfig.BaseConfig.Server.HttpPort == 0 {
		logging.Debug("[Run] No http Serve port for application ")
		return NoHttpPort
	} else {
		Serve(config.AppConfig.BaseConfig.Server.HttpPort, h.handler)
	}

	return nil
}
