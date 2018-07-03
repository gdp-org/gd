/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package config

import (
	"encoding/json"
	"flag"
	"github.com/xuyu/logging"
	"godog/utils"
)

var (
	AppConfig *DogAppConfig
)

type DogAppConfig struct {
	BaseConfig *BaseConfigure
	data       map[string]string
}

type BaseConfigure struct {
	Log struct {
		File   string
		Level  string
		Name   string
		Suffix string
	}

	Prog struct {
		CPU        int
		Daemon     bool
		HealthPort int
	}

	Server struct {
		HttpPort int
		TcpPort  int
	}
}

func init() {
	AppConfig = &DogAppConfig{
		BaseConfig: new(BaseConfigure),
		data:       make(map[string]string),
	}

	AppConfig.initNewConfigure()
}

func (a *DogAppConfig) initNewConfigure() {
	total := map[string]interface{}{}
	err := a.getConfig(a.BaseConfig, &total)
	if err != nil {
		logging.Error("[config.Go] Cannot parse config file, error = %s", err.Error())
		panic(err)
	}

	for k, v := range total {
		if s, ok := v.(string); ok {
			a.Set(k, s)
		}
	}
}

func (a *DogAppConfig) getConfig(base interface{}, appCfg interface{}) error {
	configFile := flag.String("c", "conf/conf.json", "config file pathname")
	flag.Parse()

	if appCfg == nil {
		return utils.ParseJSON(*configFile, base)
	}

	if err := utils.ParseJSON(*configFile, appCfg); err != nil {
		logging.Error("[initNewConfigure] Parse config %s. error: %s\n", *configFile, err.Error())
		return err
	}

	bytes, _ := json.Marshal(appCfg)
	_ = json.Unmarshal(bytes, base)

	return nil
}

func (a *DogAppConfig) Set(key string, value string) {
	if v, ok := a.data[key]; ok {
		logging.Warning("[dogAppConfig.Set] Try to replace value[%#+v] to key = %s, original value: %s", value, key, v)
	}

	a.data[key] = value
	logging.Info("[dogAppConfig.Set] Add/Replace [key: %s, value: %#+v] into config ok", key, value)
}

func (a *DogAppConfig) Get(key string) string {
	if v, ok := a.data[key]; ok {
		return v
	}

	logging.Error("[dogAppConfig.Get] Failed to get value of key[%s], value is NULL", key)
	return ""
}
