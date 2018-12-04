/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package godog

import (
	"github.com/chuck1024/godog/config"
	"github.com/chuck1024/godog/log"
	"github.com/chuck1024/godog/net/httplib"
	"github.com/chuck1024/godog/net/tcplib"
	"github.com/chuck1024/godog/utils"
	"runtime"
	"time"
)

var (
	AppConfig    = config.AppConfig
	AppHttp      = httplib.AppHttp
	AppTcp       = tcplib.AppTcp
	AppTcpClient *tcplib.TcpClient
)

// timeout Millisecond
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
	Info("[Run] start")
	// register signal
	Signal(AppTcp)

	// init log
	log.InitLog(AppConfig.BaseConfig.Log.File, AppConfig.BaseConfig.Log.Level, AppConfig.BaseConfig.Server.AppName, AppConfig.BaseConfig.Log.Suffix, AppConfig.BaseConfig.Log.Stdout)

	// dump when error occurs
	file, err := utils.Dump(AppConfig.BaseConfig.Server.AppName)
	if err != nil {
		Error("[Run] Error occurs when initialize dump dumpPanic file, error = %s", err.Error())
	}

	// output exit info
	defer func() {
		Info("[Run] server stop...code: %d", runtime.NumGoroutine())
		time.Sleep(time.Second)
		Info("[Run] server stop...ok")
		if err := utils.ReviewDumpPanic(file); err != nil {
			Error("[Run] Failed to review dump dumpPanic file, error = %s", err.Error())
		}
	}()

	// init cpu
	err = initCPU()
	if err != nil {
		Error("[Run] Cannot init Cpu module, error = %s", err.Error())
		return err
	}

	// register handler
	AppHttp.Register()

	// http run
	healthPort := AppConfig.BaseConfig.Prog.HealthPort
	httpPort := AppConfig.BaseConfig.Server.HttpPort
	if err = AppHttp.Run(healthPort, httpPort); err != nil {
		if err == httplib.NoHttpPort {
			Info("[Run] Hasn't http server port")
		} else {
			Error("[Run] Http server occur error in running application, error = %s", err.Error())
			return err
		}
	}

	// tcp server
	tcpPort := AppConfig.BaseConfig.Server.TcpPort
	if err = AppTcp.Run(tcpPort); err != nil {
		if err == tcplib.NoTcpPort {
			Info("[Run] Hasn't tcp server port")
		} else {
			Error("[Run] Tcp server occur error in running application, error = %s", err.Error())
			return err
		}
	}

	<-Running
	return nil
}
