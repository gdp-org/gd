/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package gd

import (
	"fmt"
	"github.com/chuck1024/gd/config"
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/net/dhttp"
	"github.com/chuck1024/gd/net/dogrpc"
	"github.com/chuck1024/gd/runtime/helper"
	"github.com/chuck1024/gd/runtime/pc"
	"github.com/chuck1024/gd/runtime/stat"
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

	enable := e.Config("Log", "enable").MustBool(false)
	if enable {
		port := e.Config("Server", "httpPort").MustInt()
		if port == 0 {
			port = e.Config("Server", "rcpPort").MustInt()
		}

		if err := RestoreLogConfig("", e.Config("Server", "serverName").String(),
			port, e.Config("Log", "level").String(), e.Config("Log", "logDir").String()); err != nil {

		}
		dlog.LoadConfiguration(logConfigFile)
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
	logDir := e.Config("Log", "logDir").String()
	file, err := utls.Dump(logDir, e.Config("Server", "serverName").String())
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

	// init falcon
	falconEnable := e.Config("Statistics", "falcon").MustBool(false)
	if falconEnable {
		pc.Init()
		defer pc.ClosePerfCounter()
	}

	// init stat
	statEnable := e.Config("Statistics", "stat").MustBool(false)
	if statEnable {
		statInterval := e.Config("Statistics", "statInterval").MustInt64(5)
		statFile := "stat.log"
		if logDir != "" {
			statFile = logDir + "/stat.log"
		}
		stat.StatMgrInstance().Init(statFile, time.Second*time.Duration(statInterval))
	}

	// http server
	httpPort := e.Config("Server", "httpPort").MustInt()
	if httpPort > 0 {
		dlog.Info("http server try listen port:%d", httpPort)

		e.HttpServer.HttpServerRunHost = fmt.Sprintf(":%d", httpPort)
		if err = e.HttpServer.Run(); err != nil {
			dlog.Error("Http server occur error in running application, error = %s", err.Error())
			return err
		}
		defer e.HttpServer.Stop()

		if falconEnable {
			pc.SetRunPort(httpPort)
		}
	}

	// rpc server
	rpcPort := e.Config("Server", "rcpPort").MustInt()
	if rpcPort > 0 {
		dlog.Info("rpc server try listen port:%d", rpcPort)

		if err = e.RpcServer.Run(rpcPort); err != nil {
			dlog.Error("rpc server occur error in running application, error = %s", err.Error())
			return err
		}
		defer e.RpcServer.Stop()
	}

	// health
	healthPort := e.Config("Process", "healthPort").MustInt()
	if healthPort > 0 {
		dlog.Info("health server try listen port:%d", healthPort)

		host := fmt.Sprintf(":%d", healthPort)
		health := &helper.Helper{Host: host}
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
	maxCPU := e.Config("Process", "maxCPU").MustInt()
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

	if e.Config("Process", "maxMemory").String() != "" {
		maxMemory, err := utls.ParseMemorySize(e.Config("Process", "maxMemory").String())
		if err != nil {
			dlog.Crash(fmt.Sprintf("conf field illgeal, max_memory:%s, error:%s", e.Config("Process", "maxMemory").String(), err.Error()))
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

func (e *Engine) SetConfig(name, key, value string) {
	e.conf.Section(name).Key(key).SetValue(value)
}

func (e *Engine) SetHttpServer(initer dhttp.HttpServerIniter) {
	e.HttpServer.SetInit(initer)
}

// timeout Millisecond
func NewRpcClient(timeout time.Duration, retryNum uint32) *dogrpc.RpcClient {
	client := dogrpc.NewClient(timeout, retryNum)
	return client
}

func NewHttpClient(Timeout time.Duration, Domain string) *dhttp.HttpClient {
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
