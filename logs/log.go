/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package logs

import (
	"github.com/xuyu/logging"
	"godog/config"
	"os"
	"path/filepath"
)

func init() {
	initLogger(config.AppConfig.BaseConfig.Log.File, config.AppConfig.BaseConfig.Log.Level, config.AppConfig.BaseConfig.Log.Name, config.AppConfig.BaseConfig.Log.Suffix, config.AppConfig.BaseConfig.Prog.Daemon)
}

func initLogger(logFile, logLevel, name, suffix string, daemon bool) error {
	logFile, _ = filepath.Abs(logFile)
	if err := os.MkdirAll(filepath.Dir(logFile), os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	handler, err := logging.NewTimeRotationHandler(logFile, suffix)
	if err != nil {
		return err
	}

	handler.SetLevelString(logLevel)
	handler.SetFormat(func(name, timeString string, rd *logging.Record) string {
		return "[" + timeString + "] " + name + " " + rd.Level.String() + " " + rd.Message + "\n"
	})

	logging.AddHandler(name, handler)

	if daemon {
		logging.DisableStdout()
	}

	return nil
}
