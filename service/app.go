/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package service

import (
	"errors"
	"github.com/xuyu/logging"
	"godog/config"
	_ "godog/logs"
	"godog/net/httplib"
	"godog/utils"
	"runtime"
	"time"
)

type InitHandlerFunc func() error
type Handler httplib.Handler

var (
	App *Application
)

type Application struct {
	appName     string
	AppConfig   *config.DogAppConfig
	health      Handler
	handler     Handler
	initHandler InitHandlerFunc
	handlerMap  map[string]httplib.HandlerFunc
}

func NewApplication(name string) *Application {
	App = &Application{
		appName:     name,
		AppConfig:   config.AppConfig,
		health:      nil,
		handler:     nil,
		initHandler: nil,
		handlerMap:  make(map[string]httplib.HandlerFunc),
	}

	return App
}

func (app *Application) SetHealthHandler(handler Handler) {
	app.health = handler
}

func (app *Application) SetServeHandler(handler Handler) {
	app.handler = handler
}

func (app *Application) SetInitHandler(handler InitHandlerFunc) {
	app.initHandler = handler
}

func (app *Application) AddHandlerFunc(addr string, handler httplib.HandlerFunc) {
	_, ok := app.handlerMap[addr]
	if ok {
		logging.Warning("[App.AddHandlerFunc] Try to replace handler to addr = %s", addr)
	}

	app.handlerMap[addr] = handler
	logging.Info("[App.AddHandlerFunc] Add/Replace [addr: %s] ok", addr)
}

func (app *Application) initCPU() error {
	if config.AppConfig.BaseConfig.Prog.CPU == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU()) //配0就用所有核
	} else {
		runtime.GOMAXPROCS(config.AppConfig.BaseConfig.Prog.CPU)
	}

	return nil
}

func (app *Application) register() {
	for k, v := range app.handlerMap {
		httplib.HandleFunc(k, v)
		logging.Info("[App.register] register handler[addr: %s]", k)
	}
}

func (app *Application) run() error {
	// health
	if app.AppConfig.BaseConfig.Prog.HealthPort != "" && app.health != nil {
		httplib.Health(app.AppConfig.BaseConfig.Prog.HealthPort, app.health)
	}

	// service
	if app.AppConfig.BaseConfig.Server.PortInfo == "" {
		return errors.New("Invalid Serve port for application ")
	}

	httplib.Serve(app.AppConfig.BaseConfig.Server.PortInfo, app.handler)

	return nil
}

func (app *Application) Run() error {
	logging.Info("[App.Run] start")
	// register signal
	utils.Signal()

	// dump when error occurs
	file, err := utils.Dump(app.appName)
	if err != nil {
		logging.Error("[App.Run] Error occurs when initialize dump panic file, error = %s", err.Error())
	}

	// output exit info
	defer func() {
		logging.Info("[App.Run] server stop...code: %d", runtime.NumGoroutine())
		time.Sleep(time.Second)
		logging.Info("[App.Run] server stop...ok")
		if err := utils.ReviewDumpPanic(file); err != nil {
			logging.Error("[App.Run] Failed to review dump panic file, error = %s", err.Error())
		}
	}()

	// init cpu
	err = app.initCPU()
	if err != nil {
		logging.Error("[App.Run] Cannot init Cpu module, error = %s", err.Error())
		return err
	}

	if app.initHandler != nil {
		err := app.initHandler()
		if err != nil {
			logging.Error("[App.Run] Error occurs when initialize application, error = %s", err.Error())
			return err
		}
	}

	// register handler
	app.register()

	// run
	err = app.run()
	if err != nil {
		logging.Error("[App.Run] Error in running application, error = %s", err.Error())
		return err
	}

	<-utils.Running
	return err
}
