/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package godog

import (
	"fmt"
	"github.com/chuck1024/doglog"
	"github.com/chuck1024/godog/config"
	"github.com/chuck1024/godog/net/dogrpc"
	"github.com/chuck1024/godog/net/httplib"
	"github.com/chuck1024/godog/utils"
	"runtime"
	"syscall"
	"time"
)

type Engine struct {
	Config     *config.DogAppConfig
	HttpServer *httplib.HttpServer
	RpcServer  *dogrpc.RpcServer
}

func Default() *Engine {
	return &Engine{
		Config: config.NewDogConfig(),
		HttpServer: &httplib.HttpServer{
			NoGinLog: true,
		},
		RpcServer: dogrpc.NewDogRpcServer(),
	}
}

// timeout Millisecond
func (e *Engine) NewRpcClient(timeout time.Duration, retryNum uint32) *dogrpc.RpcClient {
	client := dogrpc.NewClient(timeout, retryNum)
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

func (e *Engine) SetHttpServer(initer httplib.HttpServerIniter) {
	e.HttpServer.SetInit(initer)
}

func (e *Engine) initCPUAndMemory() error {
	maxCPU := e.Config.BaseConfig.Prog.MaxCPU
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

	if e.Config.BaseConfig.Prog.MaxMemory != "" {
		maxMemory, err := utils.ParseMemorySize(e.Config.BaseConfig.Prog.MaxMemory)
		if err != nil {
			doglog.Crash(fmt.Sprintf("conf field illgeal, max_memory:%s, error:%s", e.Config.BaseConfig.Prog.MaxMemory, err.Error()))
		}

		var rlimit syscall.Rlimit
		syscall.Getrlimit(syscall.RLIMIT_AS, &rlimit)
		doglog.Info("old rlimit mem:%v", rlimit)
		rlimit.Cur = uint64(maxMemory)
		rlimit.Max = uint64(maxMemory)
		err = syscall.Setrlimit(syscall.RLIMIT_AS, &rlimit)
		if err != nil {
			doglog.Crash(fmt.Sprintf("syscall Setrlimit fail, rlimit:%v, error:%s", rlimit, err.Error()))
		} else {
			syscall.Getrlimit(syscall.RLIMIT_AS, &rlimit)
			doglog.Info("new rlimit mem:%v", rlimit)
		}
	}

	return nil
}

func (e *Engine) InitLog() {
	// init log
	if e.Config.BaseConfig.Log != "" {
		doglog.LoadConfiguration(e.Config.BaseConfig.Log)
		doglog.Info("config:%v", e.Config)
	}
}

func (e *Engine) Run() error {
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

	// init cpu and memory
	err = e.initCPUAndMemory()
	if err != nil {
		doglog.Error("[Run] Cannot init CPU and memory module, error = %s", err.Error())
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

	// rpc server
	rpcPort := e.Config.BaseConfig.Server.RpcPort
	if rpcPort == 0 {
		doglog.Info("[Run] Hasn't rpc server port")
	} else {
		if err = e.RpcServer.Run(rpcPort); err != nil {
			doglog.Error("[Run] rpc server occur error in running application, error = %s", err.Error())
			return err
		}
	}

	<-Running
	return nil
}
