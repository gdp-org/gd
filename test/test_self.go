/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/xuyu/logging"
	me "godog/error"
	"godog/net/httplib"
	"godog/service"
	"net/http"
)

var Apps *service.Application

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

	var merr *me.MError
	req := &test{}
	resp := ""

	// defer write response
	defer func() {
		if merr != nil {
			logging.Error("test, errorCode: %d, errMsg: %s", merr.Code(), merr.Detail())
		}

		w.Write(httplib.LogGetResponseInfo(r, merr, resp))
	}()

	// get request data
	err := httplib.GetRequestBody(r, &req)
	if err != nil {
		merr = me.MakeHttpError(me.ERR_CODE_PARA_ERROR, err)
		return
	}
	logging.Info("test recv request: %#v", req)

	// response data
	resp = "test success!!!"
}

func register() {
	Apps.AddHandlerFunc("/test/self", HandlerTestSelf)
}

func main() {
	Apps = service.NewApplication("test")
	register()

	err := Apps.Run()
	if err != nil {
		logging.Error("Error occurs, error = %s", err.Error())
		return
	}
}
