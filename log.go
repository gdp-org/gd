/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package gd

import (
	"encoding/xml"
	"fmt"
	"github.com/chuck1024/gd/utls"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	l             = sync.Mutex{}
	logConfigFile = "conf/log.xml"
	defaultLogDir = "log"
	defaultFormat = "%L	%D %T	%l	%I	%G	%M	%S"
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

func getInfoFileName(binName string, port int) string {
	if port == 0 {
		return fmt.Sprintf("%s.log", binName)
	}
	return fmt.Sprintf("%s_%d.log", binName, port)
}

func getWarnFileName(binName string, port int) string {
	if port == 0 {
		return fmt.Sprintf("%s_err.log", binName)
	}
	return fmt.Sprintf("%s_err_%d.log", binName, port)
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
