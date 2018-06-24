/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/xuyu/logging"
	"godog/service"
	"net/http"
)

var App *service.Application

func HandlerTest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test success!!!"))
}

func main() {
	App = service.NewApplication(App.AppConfig.BaseConfig.Log.Name)
	App.AddHandlerFunc("/test", HandlerTest)

	err := App.Run()
	if err != nil {
		logging.Error("Error occurs, error = %s", err.Error())
		return
	}
}
