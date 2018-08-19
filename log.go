/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package godog

import (
	_ "github.com/chuck1024/godog/log"
	"github.com/xuyu/logging"
)

func Debug(format string, values ...interface{}) {
	logging.DefaultLogger.Log(logging.DEBUG, format, values...)
}

func Info(format string, values ...interface{}) {
	logging.DefaultLogger.Log(logging.INFO, format, values...)
}

func Warning(format string, values ...interface{}) {
	logging.DefaultLogger.Log(logging.WARNING, format, values...)
}

func Error(format string, values ...interface{}) {
	logging.DefaultLogger.Log(logging.ERROR, format, values...)
}

func ResetLogLevel(level string) {
	logging.DefaultLogger.ResetLogLevel(level)
}
