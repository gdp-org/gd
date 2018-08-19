/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package godog

import (
	"github.com/chuck1024/godog/config"
	"github.com/chuck1024/godog/dumpPanic"
	_ "github.com/chuck1024/godog/log"
	"github.com/chuck1024/godog/net/httplib"
	"github.com/chuck1024/godog/net/tcplib"
	"runtime"
	"time"
)

var (
	AppConfig    *config.DogAppConfig
	AppHttp      *httplib.HttpServer
	AppTcp       *tcplib.TcpServer
	AppTcpClient *tcplib.TcpClient
)

func init() {
	AppConfig = config.AppConfig
	AppHttp = httplib.AppHttp
	AppTcp = tcplib.AppTcp
}

func NewTcpClient(timeout, retryNum uint32) *tcplib.TcpClient {
	AppTcpClient = tcplib.NewClient(timeout, retryNum)
	return AppTcpClient
}

func initCPU() error {
	if AppConfig.BaseConfig.Prog.CPU == httplib.NoPort {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		runtime.GOMAXPROCS(AppConfig.BaseConfig.Prog.CPU)
	}

	return nil
}

func Run() error {
	Info("[App.Run] start")
	// register signal
	dumpPanic.Signal()

	// dump when error occurs
	file, err := dumpPanic.Dump(AppConfig.BaseConfig.Server.AppName)
	if err != nil {
		Error("[App.Run] Error occurs when initialize dump dumpPanic file, error = %s", err.Error())
	}

	// output exit info
	defer func() {
		Info("[App.Run] server stop...code: %d", runtime.NumGoroutine())
		time.Sleep(time.Second)
		Info("[App.Run] server stop...ok")
		if err := dumpPanic.ReviewDumpPanic(file); err != nil {
			Error("[App.Run] Failed to review dump dumpPanic file, error = %s", err.Error())
		}
	}()

	// init cpu
	err = initCPU()
	if err != nil {
		Error("[App.Run] Cannot init Cpu module, error = %s", err.Error())
		return err
	}

	// register handler
	AppHttp.Register()

	// http run
	if err = AppHttp.Run(); err != nil {
		if err == httplib.NoHttpPort {
			Info("[App.Run] Hasn't http server port")
		} else {
			Error("[App.Run] Http server occur error in running application, error = %s", err.Error())
			return err
		}
	}

	// tcp server
	if err = AppTcp.Run(); err != nil {
		if err == tcplib.NoTcpPort {
			Info("[App.Run] Hasn't tcp server port")
		} else {
			Error("[App.Run] Tcp server occur error in running application, error = %s", err.Error())
			return err
		}
	}

	<-dumpPanic.Running
	return nil
}
