/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"godog"
	"godog/config"
	_ "godog/log" // init log
)

var AppConfig *config.DogAppConfig

func main() {
	AppConfig = config.AppConfig

	// Notice: config contains BaseConfigure. config.json must contain the BaseConfigure configuration.
	// The location of config.json is "conf/conf.json". Of course, you change it if you want.

	// AppConfig.BaseConfig.Log.File is the path of log file.
	file := AppConfig.BaseConfig.Log.File
	godog.Debug("log file:%s", file)

	// AppConfig.BaseConfig.Log.Level is log level.
	// DEBUG   logLevel = 1
	// INFO    logLevel = 2
	// WARNING logLevel = 3
	// ERROR   logLevel = 4
	// DISABLE logLevel = 255
	level := AppConfig.BaseConfig.Log.Level
	godog.Debug("log level:%s", level)

	// AppConfig.BaseConfig.Log.Name is service name
	name := AppConfig.BaseConfig.Log.Name
	godog.Debug("name:%s", name)

	// AppConfig.BaseConfig.Log.Suffix is suffix of log file.
	// suffix = "060102-15" . It indicates that the log is cut per hour
	// suffix = "060102" . It indicates that the log is cut per day
	suffix := AppConfig.BaseConfig.Log.Suffix
	godog.Debug("log suffix:%s", suffix)

	// you can add configuration items directly in conf.json
	stringValue, err := AppConfig.String("stringKey")
	if err != nil {
		godog.Error("get key occur error: %s", err)
		return
	}
	godog.Debug("value:%s", stringValue)

	intValue, err := AppConfig.Int("intKey")
	if err != nil {
		godog.Error("get key occur error: %s", err)
		return
	}
	godog.Debug("value:%d", intValue)

	BoolValue, err := AppConfig.Bool("boolKey")
	if err != nil {
		godog.Error("get key occur error: %s", err)
		return
	}
	godog.Debug("value:%t", BoolValue)

	// you can add config key-value if you need.
	AppConfig.Set("yourKey", "yourValue")

	// get config key
	yourValue, err := AppConfig.String("yourKey")
	if err != nil {
		godog.Error("get key occur error: %s", err)
		return
	}
	godog.Debug("yourValue:%s", yourValue)
}
