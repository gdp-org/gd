/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"godog"
	me "godog/error"
	"godog/net/httplib"
	"net/http"
)

var Apps *godog.Application

type test struct {
	Data string
}

func HandlerTestSelf(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", httplib.CONTENT_ALL)
	w.Header().Add("Content-Type", httplib.CONTENT_JSON)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	} else if r.Method != http.MethodPost {
		// only support POST method
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var merr *me.CodeError
	req := &test{}
	resp := ""

	// defer write response
	defer func() {
		if merr != nil {
			godog.Error("test, errorCode: %d, errMsg: %s", merr.Code(), merr.Detail())
		}

		w.Write(httplib.LogGetResponseInfo(r, merr, resp))
	}()

	// get request data
	err := httplib.GetRequestBody(r, &req)
	if err != nil {
		merr = me.MakeCodeError(me.ParameterError, err)
		return
	}
	godog.Info("test recv request: %#v", req)

	// response data
	resp = "test success!!!"
}

func HandlerTcpTestSelf(req []byte) (uint16, []byte) {
	godog.Debug("tcp server request: %s", string(req))
	code := uint16(0)
	resp := []byte("Are you ok")
	return code, resp
}

func register() {
	// http
	Apps.AppHttp.AddHandlerFunc("/test/self", HandlerTestSelf)
	// Tcp
	Apps.AppTcpServer.AddTcpHandler(1024, HandlerTcpTestSelf)
}

func main() {
	Apps = godog.NewApplication("test")
	register()

	err := Apps.Run()
	if err != nil {
		godog.Error("Error occurs, error = %s", err.Error())
		return
	}
}
