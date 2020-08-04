/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package gd

import (
	"fmt"
	"github.com/chuck1024/dlog"
	"github.com/chuck1024/gd/config"
	"github.com/chuck1024/gd/net/dhttp"
	"github.com/chuck1024/gd/net/dogrpc"
	"github.com/chuck1024/gd/utls"
	"gopkg.in/ini.v1"
	"runtime"
	"syscall"
	"time"
)

type Engine struct {
	conf       *config.Conf
	HttpServer *dhttp.HttpServer
	RpcServer  *dogrpc.RpcServer
}

func Default() *Engine {
	e := &Engine{
		conf: config.Config(),
		HttpServer: &dhttp.HttpServer{
			NoGinLog: true,
		},
		RpcServer: dogrpc.NewDogRpcServer(),
	}

	if e.Config("Log","log").String() != "" {
		httpPort, _ := e.Config("Server","httpPort").Int()
		if err := RestoreLogConfig("", e.Config("Server","serverName").String(),
			httpPort, e.Config("Log","level").String(), e.Config("Log","logDir").String()); err != nil {

		}
		dlog.LoadConfiguration(logConfigFile)
		dlog.Info("config:%v", e.Config)
	}

	return e
}

// Engine Run
func (e *Engine) Run() error {
	dlog.Info("- - - - - - - - - - - - - - - - - - -")
	dlog.Info("process start")
	// register signal
	e.Signal()

	// dump when error occurs
	file, err := utls.Dump(e.Config("Server","serverName").String())
	if err != nil {
		dlog.Error("Error occurs when initialize dump dumpPanic file, error = %s", err.Error())
	}

	// output exit info
	defer func() {
		dlog.Info("server stop...code: %d", runtime.NumGoroutine())
		time.Sleep(time.Second)
		dlog.Info("server stop...ok")
		dlog.Info("- - - - - - - - - - - - - - - - - - -")
		if err := utls.ReviewDumpPanic(file); err != nil {
			dlog.Error("Failed to review dump dumpPanic file, error = %s", err.Error())
		}
	}()

	// init cpu and memory
	err = e.initCPUAndMemory()
	if err != nil {
		dlog.Error("Cannot init CPU and memory module, error = %s", err.Error())
		return err
	}

	// http run
	httpPort, _ := e.Config("Server","httpPort").Int()
	if httpPort == 0 {
		dlog.Info("Hasn't http server port")
	} else {
		e.HttpServer.HttpServerRunHost = fmt.Sprintf(":%d", httpPort)
		if err = e.HttpServer.Run(); err != nil {
			dlog.Error("Http server occur error in running application, error = %s", err.Error())
			return err
		}
		defer e.HttpServer.Stop()
	}

	// rpc server
	rpcPort, _ := e.Config("Server","rcpPort").Int()
	if rpcPort == 0 {
		dlog.Info("Hasn't rpc server port")
	} else {
		if err = e.RpcServer.Run(rpcPort); err != nil {
			dlog.Error("rpc server occur error in running application, error = %s", err.Error())
			return err
		}
		defer e.RpcServer.Stop()
	}

	// health port
	healthPort, _ := e.Config("Process","healthPort").Int()
	if healthPort == 0 {
		dlog.Info("Hasn't health server port")
	} else {
		host := fmt.Sprintf(":%d", healthPort)
		health := &utls.Helper{Host: host}
		if err := health.Start(); err != nil {
			dlog.Error("start health failed on %s\n", host)
			return err
		}
		defer health.Close()
	}

	<-Running
	return nil
}

func (e *Engine) initCPUAndMemory() error {
	maxCPU, _ := e.Config("Process","maxCPU").Int()
	numCpus := runtime.NumCPU()
	if maxCPU <= 0 {
		if numCpus > 3 {
			maxCPU = numCpus / 2
		} else {
			maxCPU = 1
		}
	} else if maxCPU > numCpus {
		maxCPU = numCpus
	}
	runtime.GOMAXPROCS(maxCPU)

	if e.Config("Process","maxMemory").String() != "" {
		maxMemory, err := utls.ParseMemorySize(e.Config("Process","maxMemory").String())
		if err != nil {
			dlog.Crash(fmt.Sprintf("conf field illgeal, max_memory:%s, error:%s", e.Config("Process","maxMemory").String(), err.Error()))
		}

		var rlimit syscall.Rlimit
		syscall.Getrlimit(syscall.RLIMIT_AS, &rlimit)
		dlog.Info("old rlimit mem:%v", rlimit)
		rlimit.Cur = uint64(maxMemory)
		rlimit.Max = uint64(maxMemory)
		err = syscall.Setrlimit(syscall.RLIMIT_AS, &rlimit)
		if err != nil {
			dlog.Crash(fmt.Sprintf("syscall Setrlimit fail, rlimit:%v, error:%s", rlimit, err.Error()))
		} else {
			syscall.Getrlimit(syscall.RLIMIT_AS, &rlimit)
			dlog.Info("new rlimit mem:%v", rlimit)
		}
	}

	return nil
}

func (e *Engine) Config(name, key string) *ini.Key {
	return e.conf.Section(name).Key(key)
}

// timeout Millisecond
func (e *Engine) NewRpcClient(timeout time.Duration, retryNum uint32) *dogrpc.RpcClient {
	client := dogrpc.NewClient(timeout, retryNum)
	return client
}

func (e *Engine) NewHttpClient(Timeout time.Duration, Domain string) *dhttp.HttpClient {
	client := &dhttp.HttpClient{
		Timeout: Timeout,
		Domain:  Domain,
	}
	if err := client.Start(); err != nil {
		dlog.Error("http client start occur error:%s", err.Error())
		return nil
	}
	return client
}

func (e *Engine) SetHttpServer(initer dhttp.HttpServerIniter) {
	e.HttpServer.SetInit(initer)
}
