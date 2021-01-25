/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package gd

import (
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
	signal.Notify(Shutdown, syscall.SIGINT, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR2)
	signal.Notify(Hup, syscall.SIGHUP)
}

func (e *Engine) Signal() {
	go func() {
		for {
			select {
			case sig := <-Shutdown:
				Info("receive signal: %v, to stop server...", sig)
				Running <- false
			case <-Hup:
			}
		}
	}()
	Info("register signal ok")
}
