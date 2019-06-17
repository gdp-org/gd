/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package godog

import (
	"fmt"
	"github.com/chuck1024/doglog"
	"github.com/chuck1024/godog/config"
	"github.com/chuck1024/godog/net/httplib"
	"github.com/chuck1024/godog/net/tcplib"
	"github.com/chuck1024/godog/utils"
	"runtime"
	"time"
)

type Engine struct {
	Config     *config.DogAppConfig
	HttpServer *httplib.HttpServer
	TcpServer  *tcplib.TcpServer
}

func Default() *Engine {
	return &Engine{
		Config: config.NewDogConfig(),
		HttpServer: &httplib.HttpServer{
			NoGinLog: true,
		},
		TcpServer: tcplib.NewTcpServer(),
	}
}

// timeout Millisecond
func (e *Engine) NewTcpClient(timeout time.Duration, retryNum uint32) *tcplib.TcpClient {
	client := tcplib.NewClient(timeout, retryNum)
	return client
}

func (e *Engine) NewHttpClient(Timeout time.Duration, Domain string) *httplib.HttpClient {
	client := &httplib.HttpClient{
		Timeout: Timeout,
		Domain:  Domain,
	}
	if err := client.Start(); err != nil {
		doglog.Error("[NewHttpClient] http client start occur error:%s", err.Error())
		return nil
	}
	return client
}

func (e *Engine) NewHttpServer(initer httplib.HttpServerIniter) {
	e.HttpServer.SetInit(initer)
}

func (e *Engine) initCPU() error {
	if e.Config.BaseConfig.Prog.CPU == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		runtime.GOMAXPROCS(e.Config.BaseConfig.Prog.CPU)
	}

	return nil
}

func (e *Engine) Run() error {
	// init log
	if e.Config.BaseConfig.Log != "" {
		doglog.LoadConfiguration(e.Config.BaseConfig.Log)
	}

	doglog.Info("[Run] start")
	// register signal
	e.Signal()

	// dump when error occurs
	file, err := utils.Dump(e.Config.BaseConfig.Server.AppName)
	if err != nil {
		doglog.Error("[Run] Error occurs when initialize dump dumpPanic file, error = %s", err.Error())
	}

	// output exit info
	defer func() {
		doglog.Info("[Run] server stop...code: %d", runtime.NumGoroutine())
		time.Sleep(time.Second)
		doglog.Info("[Run] server stop...ok")
		if err := utils.ReviewDumpPanic(file); err != nil {
			doglog.Error("[Run] Failed to review dump dumpPanic file, error = %s", err.Error())
		}
	}()

	// init cpu
	err = e.initCPU()
	if err != nil {
		doglog.Error("[Run] Cannot init CPU module, error = %s", err.Error())
		return err
	}

	// http run
	if e.Config.BaseConfig.Server.HttpPort == 0 {
		doglog.Info("[Run] Hasn't http server port")
	} else {
		e.HttpServer.HttpServerRunHost = fmt.Sprintf(":%d", e.Config.BaseConfig.Server.HttpPort)
		if err = e.HttpServer.Run(); err != nil {
			doglog.Error("[Run] Http server occur error in running application, error = %s", err.Error())
			return err
		}
	}

	// tcp server
	tcpPort := e.Config.BaseConfig.Server.TcpPort
	if err = e.TcpServer.Run(tcpPort); err != nil {
		if err == tcplib.NoTcpPort {
			doglog.Info("[Run] Hasn't tcp server port")
		} else {
			doglog.Error("[Run] Tcp server occur error in running application, error = %s", err.Error())
			return err
		}
	}

	<-Running
	return nil
}
