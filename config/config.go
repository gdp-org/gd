/**
 * Copyright 2020 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package config

import (
	"github.com/chuck1024/doglog"
	"gopkg.in/ini.v1"
	"sync"
)

var (
	defaultConfigName = "conf/conf.ini"
	cache             sync.Map
)

type Conf struct {
	ini *ini.File
}

func (c *Conf) Section(name string) *ini.Section {
	return c.ini.Section(name)
}

func Config() *Conf {
	cfg, ok := getFile(defaultConfigName)
	if !ok {
		tmp, err := ini.Load(defaultConfigName)
		if err != nil {
			doglog.Crash("Config ini load occur error:%v", err)
			return nil
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
