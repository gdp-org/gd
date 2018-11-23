/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package godog

import (
	"github.com/chuck1024/godog/net/tcplib"
	"github.com/xuyu/logging"
	"os"
	"os/signal"
	"syscall"
)

var (
	Shutdown = make(chan os.Signal)
	Running  = make(chan bool)
	Hup      = make(chan os.Signal)
)

func init() {
	signal.Notify(Shutdown, syscall.SIGINT)
	signal.Notify(Shutdown, syscall.SIGTERM)
	signal.Notify(Hup, syscall.SIGHUP)
}

func Signal(AppTcp *tcplib.TcpServer) {
	go func() {
		for {
			select {
			case <-Shutdown:
				logging.Info("[Signal] receive signal SIGINT or SIGTERM, to stop server...")
				//if config.AppConfig.BaseConfig.Server.TcpPort != httplib.NoPort {
				if AppTcp.GetAddr() != "" {
					AppTcp.Stop()
				}
				Running <- false
			case <-Hup:
			}
		}
	}()
	logging.Info("[Signal] register signal ok")
}
