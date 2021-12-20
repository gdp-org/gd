/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package gd

import (
	"github.com/chuck1024/gd/config"
	"gopkg.in/ini.v1"
)

// set conf path
func SetConfPath(path string) {
	config.SetConfPath(path)
}

func GetConfFile() *ini.File {
	return config.Config().GetIniFile()
}

// get config
func Config(name, key string) *ini.Key {
	return config.Config().Section(name).Key(key)
}

// set config
func SetConfig(name, key, value string) {
	config.Config().Section(name).Key(key).SetValue(value)
}
