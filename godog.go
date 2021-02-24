/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package gd

import (
	"fmt"
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/net/dgrpc"
	"github.com/chuck1024/gd/net/dhttp"
	"github.com/chuck1024/gd/net/dogrpc"
	"github.com/chuck1024/gd/runtime/helper"
	"github.com/chuck1024/gd/runtime/inject"
	"github.com/chuck1024/gd/runtime/pc"
	"github.com/chuck1024/gd/runtime/stat"
	"github.com/chuck1024/gd/utls"
	"google.golang.org/grpc"
	"os"
	"runtime"
	"syscall"
	"time"
)

type Engine struct {
	HttpServer *dhttp.HttpServer
	RpcServer  *dogrpc.RpcServer
	GrpcServer *dgrpc.GrpcServer
}

func Default() *Engine {
	e := &Engine{
		HttpServer: &dhttp.HttpServer{},
		RpcServer:  dogrpc.NewDogRpcServer(),
		GrpcServer: &dgrpc.GrpcServer{},
	}

	initLog()
	inject.InitDefault()
	inject.SetLogger(dlog.Global)
	return e
}

func initLog() {
	enable := Config("Log", "enable").MustBool(false)
	if enable {
		var port int
		if Config("Server", "httpPort").MustInt() > 0 {
			port = Config("Server", "httpPort").MustInt()
		} else if Config("Server", "rpcPort").MustInt() > 0 {
			port = Config("Server", "rpcPort").MustInt()
		} else if Config("Server", "grpcPort").MustInt() > 0 {
			port = Config("Server", "grpcPort").MustInt()
		}

		log := &gdConfig{
			BinName:    Config("Server", "serverName").String(),
			Port:       port,
			LogLevel:   Config("Log", "level").MustString(defaultFormat),
			LogDir:     Config("Log", "logDir").String(),
			Stdout:     Config("Log", "stdout").MustString("true"),
			Format:     Config("Log", "format").MustString(defaultFormat),
			Rotate:     Config("Log", "rotate").MustString("true"),
			Maxsize:    Config("Log", "maxsize").MustString("0M"),
			MaxLines:   Config("Log", "maxLines").MustString("0k"),
			RotateType: Config("Log", "rotateType").MustString("hourly"),
		}

		if err := log.initLogConfig(); err != nil {
			panic(fmt.Sprintf("initLogConfig occur error:%v", err))
		}
	}
}

// Engine Run
func (e *Engine) Run() error {
	Info("- - - - - - - - - - - - - - - - - - -")
	Info("process start")
	// register signal
	e.Signal()

	defer inject.Close()

	var err error
	// dump when error occurs
	logDir := Config("Log", "logDir").String()
	file := &os.File{}
	if Config("Log", "toFile").MustString("false") == "true" {
		file, err = utls.Dump(logDir, Config("Server", "serverName").String())
		if err != nil {
			Error("Error occurs when initialize dump dumpPanic file, error = %s", err.Error())
		}
	}

	// output exit info
	defer func() {
		Info("server stop...code: %d", runtime.NumGoroutine())
		time.Sleep(time.Second)
		Info("server stop...ok")
		Info("- - - - - - - - - - - - - - - - - - -")

		if Config("Log", "toFile").MustString("false") == "false" {
			return
		}

		if err := utls.ReviewDumpPanic(file); err != nil {
			Error("Failed to review dump dumpPanic file, error = %s", err.Error())
		}
	}()

	// init cpu and memory
	err = e.initCPUAndMemory()
	if err != nil {
		Error("Cannot init CPU and memory module, error = %s", err.Error())
		return err
	}

	// init falcon
	falconEnable := Config("Statistics", "falcon").MustBool(false)
	if falconEnable {
		pc.Init()
		defer pc.ClosePerfCounter()
	}

	// init stat
	statEnable := Config("Statistics", "stat").MustBool(false)
	if statEnable {
		statInterval := Config("Statistics", "statInterval").MustInt64(5)
		statFile := "stat.log"
		if Config("Log", "toFile").MustString("false") == "true" {
			if logDir != "" {
				statFile = logDir + "/stat.log"
			}
		}
		stat.StatMgrInstance().Init(statFile, time.Second*time.Duration(statInterval))
	}

	// http server
	httpPort := Config("Server", "httpPort").MustInt()
	if httpPort > 0 {
		Info("http server try listen port:%d", httpPort)
		inject.RegisterOrFail("httpServerRunHost", httpPort)
		inject.RegisterOrFail("httpServer", e.HttpServer)

		if falconEnable {
			pc.SetRunPort(httpPort)
		}
	}

	// grpc server
	grpcPort := Config("Server", "grpcPort").MustInt()
	if grpcPort > 0 {
		Info("grpc server try listen port:%d", grpcPort)
		inject.RegisterOrFail("grpcRunHost", grpcPort)
		inject.RegisterOrFail("serviceName", Config("Server", "serverName").String())
		inject.RegisterOrFail("grpcServer", e.GrpcServer)

		if falconEnable {
			pc.SetRunPort(grpcPort)
		}
	}

	// health
	healthPort := Config("Process", "healthPort").MustInt()
	if healthPort > 0 {
		Info("health server try listen port:%d", healthPort)
		inject.RegisterOrFail("helperHost", healthPort)
		inject.RegisterOrFail("helper", (*helper.Helper)(nil))
	}

	// rpc server
	rpcPort := Config("Server", "rpcPort").MustInt()
	if rpcPort > 0 {
		Info("rpc server try listen port:%d", rpcPort)

		inject.RegisterOrFail("rpcHost", rpcPort)
		inject.RegisterOrFail("rpcServer", e.RpcServer)
	}

	<-Running
	return nil
}

func (e *Engine) initCPUAndMemory() error {
	maxCPU := Config("Process", "maxCPU").MustInt()
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

	if Config("Process", "maxMemory").String() != "" {
		maxMemory, err := utls.ParseMemorySize(Config("Process", "maxMemory").String())
		if err != nil {
			Crash(fmt.Sprintf("conf field illgeal, max_memory:%s, error:%s", Config("Process", "maxMemory").String(), err.Error()))
		}

		var rlimit syscall.Rlimit
		syscall.Getrlimit(syscall.RLIMIT_AS, &rlimit)
		Info("old rlimit mem:%v", rlimit)
		rlimit.Cur = uint64(maxMemory)
		rlimit.Max = uint64(maxMemory)
		err = syscall.Setrlimit(syscall.RLIMIT_AS, &rlimit)
		if err != nil {
			Crash(fmt.Sprintf("syscall Setrlimit fail, rlimit:%v, error:%s", rlimit, err.Error()))
		} else {
			syscall.Getrlimit(syscall.RLIMIT_AS, &rlimit)
			Info("new rlimit mem:%v", rlimit)
		}
	}

	return nil
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
		Error("http client start occur error:%s", err.Error())
		return nil
	}
	return client
}

func NewGrpcClient(target string, makeRawClient func(conn *grpc.ClientConn) (interface{}, error), serviceName string) *dgrpc.GrpcClient {
	client := &dgrpc.GrpcClient{
		Target:      target,
		ServiceName: serviceName,
	}

	if err := client.Start(makeRawClient); err != nil {
		Error("grpc client start occur error:%s", err.Error())
		return nil
	}
	return client
}
