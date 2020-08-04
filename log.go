/**
 * Copyright 2020 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package godog

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/chuck1024/godog/utls"
)

var (
	l               = sync.Mutex{}
	logConfigFile   = "conf/log.xml"
	defaultLogDir   = "log"
	defaultFormat   = "%L	%D %T	%l	%I	%G	%M	%S"
	inContainerEnv  = false
)

type xmlLoggerConfig struct {
	ScribeCategory string      `xml:"scribeCategory"`
	Filter         []xmlFilter `xml:"filter"`
}

type xmlProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type xmlFilter struct {
	Enabled  string        `xml:"enabled,attr"`
	Tag      string        `xml:"tag"`
	Level    string        `xml:"level"`
	Type     string        `xml:"type"`
	Property []xmlProperty `xml:"property"`
}

func init() {
	contAppName := os.Getenv("CONTAINER_S_APPNAME")
	if contAppName != "" {
		inContainerEnv = true
	}
}

func getInfoFileName(binname string, port int) string {
	if port == 0 {
		return fmt.Sprintf("%s.log", binname)
	}
	return fmt.Sprintf("%s_%d.log", binname, port)
}

func getWarnFileName(binname string, port int) string {
	if port == 0 {
		return fmt.Sprintf("%s_err.log", binname)
	}
	return fmt.Sprintf("%s_err_%d.log", binname, port)
}

func RestoreLogConfig(configFilePath string, binName string, port int, logLevel string, logDir string) error {
	l.Lock()
	defer l.Unlock()
	if logDir == "" {
		logDir = defaultLogDir
	}

	if configFilePath == "" {
		configFilePath = logConfigFile
	}

	// redirect stdout and stderr to file
	if inContainerEnv {
		stdLogFileName := fmt.Sprintf("%s/%s", logDir, "stderr_process.log")
		if utls.PathExists(stdLogFileName) {
			//Mon Jan 2 15:04:05 -0700 MST 2006
			ct := time.Now().Format("2006-01-02-15:04:05")
			newPath := stdLogFileName + "-" + ct
			err := os.Rename(stdLogFileName, newPath)
			fmt.Println(fmt.Sprintf("rename %v => %v", stdLogFileName, newPath))
			if err != nil {
				fmt.Println(fmt.Sprintf("rename stderr file fail,f=%v,nf=%v,err=%v", stdLogFileName, newPath, err))
			}
		}

		if !utls.PathExists(stdLogFileName) {
			err := utls.EnsureDir(logDir)
			if err != nil {
				return fmt.Errorf("ensure dir fail,dir=%v,err=%w", logDir, err)
			}
			err = utls.Store2File(stdLogFileName, "")
			if err != nil {
				return fmt.Errorf("create file fail,name=%v,err=%w", stdLogFileName, err)
			}
		}

		stdLogFile, err := os.OpenFile(stdLogFileName, os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0755)
		if err != nil {
			return fmt.Errorf("open std log file fail,file=%v,err=%w", stdLogFileName, err)
		}
		err = syscall.Dup2(int(stdLogFile.Fd()), 2)
		if err != nil {
			return fmt.Errorf("redirect stderr to file fail,file=%v,err=%w", stdLogFileName, err)
		}
	}

	if utls.PathExists(configFilePath) {
		return nil
	}
	if logLevel == "" {
		logLevel = "DEBUG"
	}

	if binName == "" {
		ex, err := os.Executable()
		if err != nil {
			return err
		}
		exPath := filepath.Dir(ex)
		if strings.Contains(exPath, "/") {
			ex = ex[len(exPath)+1:]
		}
		binName = ex
	}

	if logLevel != "DEBUG" && logLevel != "INFO" && logLevel != "WARNING" && logLevel != "ERROR" {
		return fmt.Errorf("invalid log level %v", logLevel)
	}

	infoFileName := getInfoFileName(binName, port)
	warnFileName := getWarnFileName(binName, port)

	var filters []xmlFilter
	// stdout
	stdout := xmlFilter{
		Enabled: "false",
		Tag:     "stdout",
		Level:   "INFO",
		Type:    "console",
	}
	filters = append(filters, stdout)
	// info
	info := xmlFilter{
		Enabled: "true",
		Tag:     "service",
		Level:   logLevel,
		Type:    "file",
		Property: []xmlProperty{
			xmlProperty{Name: "filename", Value: fmt.Sprintf("%s/%s", logDir, infoFileName)},
			xmlProperty{Name: "format", Value: defaultFormat},
			xmlProperty{Name: "rotate", Value: "true"},
			xmlProperty{Name: "maxsize", Value: "0M"},
			xmlProperty{Name: "maxlines", Value: "0K"},
			xmlProperty{Name: "hourly", Value: "true"},
		},
	}
	filters = append(filters, info)
	// warn
	warn := xmlFilter{
		Enabled: "true",
		Tag:     "service_err",
		Level:   "WARNING",
		Type:    "file",
		Property: []xmlProperty{
			xmlProperty{Name: "filename", Value: fmt.Sprintf("%s/%s", logDir, warnFileName)},
			xmlProperty{Name: "format", Value: defaultFormat},
			xmlProperty{Name: "rotate", Value: "true"},
			xmlProperty{Name: "maxsize", Value: "0M"},
			xmlProperty{Name: "maxlines", Value: "0K"},
			xmlProperty{Name: "hourly", Value: "true"},
		},
	}
	filters = append(filters, warn)

	c := &xmlLoggerConfig{
		Filter: filters,
	}

	bts, err := xml.Marshal(c)
	if err != nil {
		return err
	}

	err = utls.Store2File(configFilePath, string(bts))
	if err != nil {
		return err
	}

	return nil
}
