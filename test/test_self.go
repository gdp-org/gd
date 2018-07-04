/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/xuyu/logging"
	me "godog/error"
	"godog/net/httplib"
	"godog/net/tcplib"
	"godog"
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

func HandlerTcpTestSelf(req tcplib.Packet) (rsp tcplib.Packet) {
	cReq := req.(*tcplib.TcpPacket)
	rsp = tcplib.NewCustomPacketWithSeq(cReq.Cmd, []byte("1024 hello."), cReq.Seq)
	return
}


func register() {
	// http
	Apps.AppHttp.AddHandlerFunc("/test/self", HandlerTestSelf)
	// Tcp
	App.AppTcpServer.AddTcpHandler(1024, HandlerTcpTestSelf)
}

func main() {
	Apps = godog.NewApplication("test")
	register()

	err := Apps.Run()
	if err != nil {
		logging.Error("Error occurs, error = %s", err.Error())
		return
	}
}
