/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main_test

import (
	"github.com/chuck1024/doglog"
	"github.com/chuck1024/godog"
	"testing"
)

func TestConfig(t *testing.T) {
	// init log
	doglog.LoadConfiguration("conf/log.xml")

	// Notice: config contains BaseConfigure. config.json must contain the BaseConfigure configuration.
	// The location of config.json is "conf/conf.json". Of course, you change it if you want.

	d := godog.Default()
	// AppConfig.BaseConfig.Log is the path of log file.
	file := d.Config.BaseConfig.Log
	t.Logf("log file:%s", file)

	// AppConfig.BaseConfig.Server.AppName is service name
	name := d.Config.BaseConfig.Server.AppName
	t.Logf("name:%s", name)

	// you can add configuration items directly in conf.json
	stringValue, err := d.Config.String("stringKey")
	if err != nil {
		doglog.Error("get key occur error: %s", err)
		return
	}
	t.Logf("value:%s", stringValue)

	stringsValue, err := d.Config.Strings("stringsKey")
	if err != nil {
		doglog.Error("get key occur error: %s", err)
		return
	}
	t.Logf("value:%s", stringsValue)

	intValue, err := d.Config.Int("intKey")
	if err != nil {
		doglog.Error("get key occur error: %s", err)
		return
	}
	t.Logf("value:%d", intValue)

	BoolValue, err := d.Config.Bool("boolKey")
	if err != nil {
		doglog.Error("get key occur error: %s", err)
		return
	}
	t.Logf("value:%t", BoolValue)

	// you can add config key-value if you need.
	d.Config.Set("yourKey", "yourValue")

	// get config key
	yourValue, err := d.Config.String("yourKey")
	if err != nil {
		doglog.Error("get key occur error: %s", err)
		return
	}
	t.Logf("yourValue:%s", yourValue)
}
