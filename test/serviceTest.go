/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"godog"
	"net/http"
)

var App *godog.Application

func HandlerHttpTest(w http.ResponseWriter, r *http.Request) {
	godog.Debug("connected : %s", r.RemoteAddr)
	w.Write([]byte("test success!!!"))
}

func HandlerTcpTest(req []byte) (uint16, []byte) {
	godog.Debug("tcp server request: %s", string(req))
	code := uint16(0)
	resp := []byte("Are you ok?")
	return code, resp
}

func main() {
	AppName := "test"
	App = godog.NewApplication(AppName)
	// Http
	App.AppHttp.AddHandlerFunc("/test", HandlerHttpTest)

	// Tcp
	App.AppTcpServer.AddTcpHandler(1024, HandlerTcpTest)

	err := App.Run()
	if err != nil {
		godog.Error("Error occurs, error = %s", err.Error())
		return
	}
}

// you can use command to test service that it is in another file <serviceTest.txt>.
