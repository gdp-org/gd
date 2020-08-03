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
	"github.com/chuck1024/godog/net/dhttp"
	"github.com/chuck1024/godog/utils"
	"runtime"
	"syscall"
	"time"
)

type Engine struct {
	Config     *config.Conf
	HttpServer *dhttp.HttpServer
	RpcServer  *dogrpc.RpcServer
}

func Default() *Engine {
	e := &Engine{
		Config: config.Config(),
		HttpServer: &dhttp.HttpServer{
			NoGinLog: true,
		},
		RpcServer: dogrpc.NewDogRpcServer(),
	}

	if e.Config.Section("Log").Key("log").String() != "" {
		doglog.LoadConfiguration(e.Config.Section("Log").Key("log").String())
		doglog.Info("config:%v", e.Config)
	}

	return e
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
		doglog.Error("http client start occur error:%s", err.Error())
		return nil
	}
	return client
}

func (e *Engine) SetHttpServer(initer dhttp.HttpServerIniter) {
	e.HttpServer.SetInit(initer)
}

func (e *Engine) initCPUAndMemory() error {
	maxCPU, _ := e.Config.Section("Process").Key("maxCPU").Int()
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

	if e.Config.Section("Process").Key("maxMemory").String() != "" {
		maxMemory, err := utils.ParseMemorySize(e.Config.Section("Process").Key("maxMemory").String())
		if err != nil {
			doglog.Crash(fmt.Sprintf("conf field illgeal, max_memory:%s, error:%s", e.Config.Section("Process").Key("maxMemory").String(), err.Error()))
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

func (e *Engine) Run() error {
	doglog.Info("- - - - - - - - - - - - - - - - - - -")
	doglog.Info("process start")
	// register signal
	e.Signal()

	// dump when error occurs
	file, err := utils.Dump(e.Config.Section("Server").Key("serverName").String())
	if err != nil {
		doglog.Error("Error occurs when initialize dump dumpPanic file, error = %s", err.Error())
	}

	// output exit info
	defer func() {
		doglog.Info("server stop...code: %d", runtime.NumGoroutine())
		time.Sleep(time.Second)
		doglog.Info("server stop...ok")
		doglog.Info("- - - - - - - - - - - - - - - - - - -")
		if err := utils.ReviewDumpPanic(file); err != nil {
			doglog.Error("Failed to review dump dumpPanic file, error = %s", err.Error())
		}
	}()

	// init cpu and memory
	err = e.initCPUAndMemory()
	if err != nil {
		doglog.Error("Cannot init CPU and memory module, error = %s", err.Error())
		return err
	}

	// http run
	httpPort, _ := e.Config.Section("Server").Key("httpPort").Int()
	if httpPort == 0 {
		doglog.Info("Hasn't http server port")
	} else {
		e.HttpServer.HttpServerRunHost = fmt.Sprintf(":%d", httpPort)
		if err = e.HttpServer.Run(); err != nil {
			doglog.Error("Http server occur error in running application, error = %s", err.Error())
			return err
		}
		defer e.HttpServer.Stop()
	}

	// rpc server
	rpcPort, _ := e.Config.Section("Server").Key("tcpPort").Int()
	if rpcPort == 0 {
		doglog.Info("Hasn't rpc server port")
	} else {
		if err = e.RpcServer.Run(rpcPort); err != nil {
			doglog.Error("rpc server occur error in running application, error = %s", err.Error())
			return err
		}
		defer e.RpcServer.Stop()
	}

	// health port
	healthPort, _ := e.Config.Section("Process").Key("healthPort").Int()
	if healthPort == 0 {
		doglog.Info("Hasn't health server port")
	} else {
		host := fmt.Sprintf(":%d", healthPort)
		health := &utils.Helper{Host: host}
		if err := health.Start(); err != nil {
			doglog.Error("start health failed on %s\n", host)
			return err
		}
		defer health.Close()
	}

	<-Running
	return nil
}
