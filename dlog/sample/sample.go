/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/gdp-org/gd/dlog"
	"github.com/gdp-org/gd/runtime/gl"
	"strconv"
	"time"
)

func main() {
	gl.Init()
	defer gl.Close()
	gl.Set(gl.LogId, strconv.FormatInt(time.Now().UnixNano(), 10))
	dlog.LoadConfiguration("log.xml")
	dlog.Debug("test:%s", "ok")
	dlog.Close()
}
