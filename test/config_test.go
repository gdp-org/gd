/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/xuyu/logging" // import logging module
	"godog/config"
	_ "godog/log" // init log
	"testing"
)

var AppConfig *config.DogAppConfig

func TestConfig(t *testing.T) {
	AppConfig = config.AppConfig

	// Notice: config contains BaseConfigure. config.json must contain the BaseConfigure configuration.
	// The location of config.json is "conf/conf.json". Of course, you change it if you want.

	// AppConfig.BaseConfig.Log.File is the path of log file.
	file := AppConfig.BaseConfig.Log.File
	logging.Debug("log file:%s", file)

	// AppConfig.BaseConfig.Log.Level is log level.
	// DEBUG   logLevel = 1
	// INFO    logLevel = 2
	// WARNING logLevel = 3
	// ERROR   logLevel = 4
	// DISABLE logLevel = 255
	level := AppConfig.BaseConfig.Log.Level
	logging.Debug("log level:%s", level)

	// AppConfig.BaseConfig.Log.Name is service name
	name := AppConfig.BaseConfig.Log.Name
	logging.Debug("name:%s", name)

	// AppConfig.BaseConfig.Log.Suffix is suffix of log file.
	// suffix = "060102-15" . It indicates that the log is cut per hour
	// suffix = "060102" . It indicates that the log is cut per day
	suffix := AppConfig.BaseConfig.Log.Suffix
	logging.Debug("log suffix:%s", suffix)

	// you can add configuration items directly in conf.json
	value := AppConfig.Get("key")
	logging.Debug("value:%s", value)

	// you can add config key-value if you need.
	AppConfig.Set("yourKey", "yourValue")

	// get config key
	yourValue := AppConfig.Get("yourKey")
	logging.Debug("yourValue:%s", yourValue)
}
