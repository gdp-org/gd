/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/xuyu/logging"
	"godog/service"
	"net/http"
	"testing"
)

var App *service.Application

func HandlerTest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test success!!!"))
}

func TestService(t *testing.T) {
	App = service.NewApplication("test")
	App.AddHandlerFunc("/test", HandlerTest)

	err := App.Run()
	if err != nil {
		logging.Error("Error occurs, error = %s", err.Error())
		return
	}
}

// you can use command to test service that it is in another file <service_test.txt>.