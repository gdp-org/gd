package main

import (
	"github.com/bitly/go-simplejson"
	"github.com/gdp-org/gd"
	"time"
)

type TestHttpClientReq struct {
	Data string
}

func main() {
	defer gd.LogClose()
	req := &TestHttpClientReq{Data: "chuck"}
	_, body, err := gd.NewHttpClient().Timeout(3 * time.Second).Post("http://127.0.0.1:10240/test").Send(req).End()
	if err != nil {
		gd.Error("occur error:%s\n", err)
		return
	}
	ret, _:= simplejson.NewJson([]byte(body))
	gd.Info(ret.Get("result").Get("Ret"))
}
