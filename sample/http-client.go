package main

import (
	"fmt"
	"github.com/chuck1024/gd"
	"github.com/chuck1024/gd/runtime/gl"
	"strconv"
	"time"
)

type TestHttpClientReq struct {
	Data string
}

func main() {
	req := &TestHttpClientReq{Data: "chuck"}
	gl.Init()
	defer gl.Close()
	gl.Set(gl.LogId, strconv.FormatInt(time.Now().UnixNano(), 10))
	_, body, err := gd.NewHttpClient().Timeout(3 * time.Second).Post("http://127.0.0.1:10240/test").Send(req).End()
	if err != nil {
		fmt.Printf("occur error:%s\n", err)
		return
	}
	fmt.Println(body)
}
