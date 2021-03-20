package main

import (
	"fmt"
	"github.com/chuck1024/gd"
)

type TestHttpClientReq struct {
	Data string
}

func main() {
	req := &TestHttpClientReq{Data: "chuck"}
	_, body, err := gd.NewHttpClient().Post("http://127.0.0.1:10240/test").Send(req).End()
	if err != nil {
		fmt.Printf("occur error:%s\n", err)
		return
	}
	fmt.Println(body)
}
