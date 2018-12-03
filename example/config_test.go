/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main_test

import (
	"github.com/chuck1024/godog"
	"github.com/chuck1024/godog/log"
	"testing"
)

func TestConfig(t *testing.T) {
	// init log
	log.InitLog(godog.AppConfig.BaseConfig.Log.File, godog.AppConfig.BaseConfig.Log.Level, godog.AppConfig.BaseConfig.Server.AppName, godog.AppConfig.BaseConfig.Log.Suffix, godog.AppConfig.BaseConfig.Log.Daemon)

	// Notice: config contains BaseConfigure. config.json must contain the BaseConfigure configuration.
	// The location of config.json is "conf/conf.json". Of course, you change it if you want.

	// AppConfig.BaseConfig.Log.File is the path of log file.
	file := godog.AppConfig.BaseConfig.Log.File
	godog.Debug("log file:%s", file)

	// AppConfig.BaseConfig.Log.Level is log level.
	// DEBUG   logLevel = 1
	// INFO    logLevel = 2
	// WARNING logLevel = 3
	// ERROR   logLevel = 4
	// DISABLE logLevel = 255
	level := godog.AppConfig.BaseConfig.Log.Level
	t.Logf("log level:%s", level)

	// AppConfig.BaseConfig.Server.AppName is service name
	name := godog.AppConfig.BaseConfig.Server.AppName
	t.Logf("name:%s", name)

	// AppConfig.BaseConfig.Log.Suffix is suffix of log file.
	// suffix = "060102-15" . It indicates that the log is cut per hour
	// suffix = "060102" . It indicates that the log is cut per day
	suffix := godog.AppConfig.BaseConfig.Log.Suffix
	t.Logf("log suffix:%s", suffix)

	// you can add configuration items directly in conf.json
	stringValue, err := godog.AppConfig.String("stringKey")
	if err != nil {
		godog.Error("get key occur error: %s", err)
		return
	}
	t.Logf("value:%s", stringValue)

	stringsValue, err := godog.AppConfig.Strings("stringsKey")
	if err != nil {
		godog.Error("get key occur error: %s", err)
		return
	}
	t.Logf("value:%s", stringsValue)

	intValue, err := godog.AppConfig.Int("intKey")
	if err != nil {
		godog.Error("get key occur error: %s", err)
		return
	}
	t.Logf("value:%d", intValue)

	BoolValue, err := godog.AppConfig.Bool("boolKey")
	if err != nil {
		godog.Error("get key occur error: %s", err)
		return
	}
	t.Logf("value:%t", BoolValue)

	// you can add config key-value if you need.
	godog.AppConfig.Set("yourKey", "yourValue")

	// get config key
	yourValue, err := godog.AppConfig.String("yourKey")
	if err != nil {
		godog.Error("get key occur error: %s", err)
		return
	}
	t.Logf("yourValue:%s", yourValue)
}
