/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package config

import (
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/utls"
	"gopkg.in/ini.v1"
	"sync"
)

var (
	defaultConfigName = "conf/conf.ini"
	cache             sync.Map
	defaultNoConfig   = &Conf{ini: ini.Empty()}
)

type Conf struct {
	ini *ini.File
}

func (c *Conf) Section(name string) *ini.Section {
	return c.ini.Section(name)
}

func (c *Conf) GetIniFile() *ini.File {
	return c.ini
}

func SetConfPath(path string) {
	if path != "" {
		defaultConfigName = path
	}
}

func Config() *Conf {
	cfg, ok := getFile(defaultConfigName)
	if !ok {
		if !utls.Exists(defaultConfigName) {
			return defaultNoConfig
		}

		tmp, err := ini.Load(defaultConfigName)
		if err != nil {
			dlog.Warn("Config ini load conf/conf.ini occur error:%v", err)
			return defaultNoConfig
		}
		setFile(defaultConfigName, tmp)
		cfg = tmp
	}
	return &Conf{ini: cfg}
}

func getFile(name string) (*ini.File, bool) {
	fo, ok := cache.Load(name)
	if !ok || fo == nil {
		return ini.Empty(), false
	}
	f, ok := fo.(*ini.File)
	if !ok || f == nil {
		return ini.Empty(), false
	}
	return f, ok
}

func setFile(name string, file *ini.File) {
	cache.Store(name, file)
}
