/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package config

import (
	"encoding/json"
	"errors"
	"github.com/chuck1024/godog/utils"
	"github.com/xuyu/logging"
	"os"
	"path/filepath"
)

var (
	AppConfig *DogAppConfig
	appConfigPath string
)

type DogAppConfig struct {
	BaseConfig *BaseConfigure
	data       map[string]interface{}
}

type BaseConfigure struct {
	Log struct {
		File   string
		Level  string
		Daemon bool
		Suffix string
	}

	Prog struct {
		CPU        int
		HealthPort int
	}

	Server struct {
		AppName  string
		HttpPort int
		TcpPort  int
	}
}

func init() {
	workPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var filename = "conf.json"
	appConfigPath = filepath.Join(workPath, "conf", filename)
	if !utils.Exists(appConfigPath) {
		AppConfig = &DogAppConfig{
			BaseConfig: new(BaseConfigure),
			data:       make(map[string]interface{}),
		}
		return
	}

	AppConfig = NewDogConfig()
}

func NewDogConfig() *DogAppConfig {
	c := &DogAppConfig{
		BaseConfig: new(BaseConfigure),
		data:       make(map[string]interface{}),
	}

	c.initNewConfigure()
	return c
}

func (a *DogAppConfig) initNewConfigure() {
	total := map[string]interface{}{}
	err := a.getConfig(a.BaseConfig, &total)
	if err != nil {
		logging.Error("[initNewConfigure] Cannot parse config file, error = %s", err.Error())
		panic(err)
	}

	for k, v := range total {
		if s, ok := v.(string); ok {
			a.Set(k, s)
		}
		if s, ok := v.(float64); ok {
			a.Set(k, s)
		}
		if s, ok := v.(bool); ok {
			a.Set(k, s)
		}
	}
}

func (a *DogAppConfig) getConfig(base interface{}, appCfg interface{}) error {
	if appCfg == nil {
		return utils.ParseJSON(appConfigPath, base)
	}

	if err := utils.ParseJSON(appConfigPath, appCfg); err != nil {
		logging.Error("[getConfig] Parse config %s. error: %s\n", appConfigPath, err.Error())
		return err
	}

	bytes, _ := json.Marshal(appCfg)
	_ = json.Unmarshal(bytes, base)

	return nil
}

func (a *DogAppConfig) Set(key string, value interface{}) {
	if v, ok := a.data[key]; ok {
		logging.Warning("[Set] Try to replace value[%#+v] to key = %s, original value: %s", value, key, v)
	}

	a.data[key] = value
}

func (a *DogAppConfig) String(key string) (string, error) {
	if v, ok := a.data[key]; ok {
		switch v.(type) {
		case string:
			return v.(string), nil
		default:
			return "", errors.New("value type isn't string")
		}
	}

	return "", errors.New("failed to get value of key. No key")
}

func (a *DogAppConfig) Int(key string) (int, error) {
	if v, ok := a.data[key]; ok {
		switch v.(type) {
		case float64:
			return int(v.(float64)), nil
		default:
			return 0, errors.New("value type isn't int")
		}
	}

	return 0, errors.New("failed to get value of key. No key")
}

func (a *DogAppConfig) Bool(key string) (bool, error) {
	if v, ok := a.data[key]; ok {
		switch v.(type) {
		case bool:
			return v.(bool), nil
		default:
			return false, errors.New("value type isn't int")
		}
	}

	return false, errors.New("failed to get value of key. No key")
}
