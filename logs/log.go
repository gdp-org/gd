/**
 * Created by JetBrains GoLand.
 * Author: Chuck Chen
 * Date: 2018/6/22
 * Time: 16:38
 */

package log

import (
	"path/filepath"
	"os"
	"github.com/xuyu/logging"
)

func InitLogger(logFile, logLevel, name, suffix string, daemon bool) error {
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