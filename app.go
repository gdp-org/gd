/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package godog

import (
	"godog/config"
	"godog/dumpPanic"
	_ "godog/log"
	"godog/net/httplib"
	"godog/net/tcplib"
	"runtime"
	"time"
)

var (
	App *Application
)

type Application struct {
	appName      string
	AppConfig    *config.DogAppConfig
	AppHttp      *httplib.HttpServer
	AppTcpServer *tcplib.TcpServer
	AppTcpClient *tcplib.TcpClient
}

func NewApplication(name string) *Application {
	App = &Application{
		appName:      name,
		AppConfig:    config.AppConfig,
		AppHttp:      httplib.AppHttp,
		AppTcpServer: tcplib.AppTcpServer,
		AppTcpClient: tcplib.AppTcpClient,
	}

	return App
}

func (app *Application) initCPU() error {
	if config.AppConfig.BaseConfig.Prog.CPU == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		runtime.GOMAXPROCS(config.AppConfig.BaseConfig.Prog.CPU)
	}

	return nil
}

func (app *Application) Run() error {
	Info("[App.Run] start")
	// register signal
	dumpPanic.Signal()

	// dump when error occurs
	file, err := dumpPanic.Dump(app.appName)
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
	err = app.initCPU()
	if err != nil {
		Error("[App.Run] Cannot init Cpu module, error = %s", err.Error())
		return err
	}

	// register handler
	app.AppHttp.Register()

	// http run
	err = app.AppHttp.Run()
	if err != nil {
		if err == httplib.NoHttpPort {
			Info("[App.Run] Hasn't http server port")
		} else {
			Error("[App.Run] Http server occur error in running application, error = %s", err.Error())
			return err
		}
	}

	// tcp server
	err = app.AppTcpServer.Run()
	if err != nil {
		if err == tcplib.NoTcpPort {
			Info("[App.Run] Hasn't tcp server port")
		} else {
			Error("[App.Run] Tcp server occur error in running application, error = %s", err.Error())
			return err
		}
	}

	<-dumpPanic.Running
	return err
}
