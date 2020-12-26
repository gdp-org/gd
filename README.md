# godog

"go" is the meaning of a dog in Chinese pronunciation, and dog's original intention is also a dog. So godog means "狗狗" in Chinese, which is very cute.

---
## Author

```
author: Chuck1024
email : chuck.ch1024@outlook.com
```

---
## Installation

Start with cloning gd:

```
> go get github.com/chuck1024/gd
```

---
## Introduction

Godog is a basic framework implemented by golang, which is aiming at helping developers setup feature-rich server quickly.

The framework contains `config module`,`databases module`,`error module`,`dlog module`,`net module`,`runtime module` and `server module`. You can select any modules according to your practice. More features will be added later. I hope anyone who is interested in this work can join it and let's enhance the system function of this framework together.

>* [gin](https://github.com/gin-gonic/gin) and [zookeeper](https://github.com/samuel/go-zookeeper) are third-party library and more third-party library in go.mod. 
>* Authors are [**Gin-Gonic**](https://gin-gonic.com/) and [**Samuel Stauffer**](https://github.com/samuel).Thanks for them here. 

---
## Quick start

```go
package main

import (
	"github.com/chuck1024/gd"
	"github.com/chuck1024/gd/net/dhttp"
	"github.com/gin-gonic/gin"
	"net/http"
)

func HandlerHttp(c *gin.Context, req interface{}) (code int, message string, err error, ret string) {
	gd.Debug("httpServerTest req:%v", req)
	ret = "ok!!!"
	return http.StatusOK, "ok", nil, ret
}

func main() {
	d := gd.Default()
	d.HttpServer.SetInit(func(g *gin.Engine) error {
		r := g.Group("")
		r.Use(
			dhttp.GlFilter(),
			dhttp.GroupFilter(),
			dhttp.Logger("quick-start"),
		)

		d.HttpServer.GET(r, "test", HandlerHttp)

		if err := d.HttpServer.CheckHandle(); err != nil {
			return err
		}
		return nil
	})

	gd.SetConfig("Server", "httpPort", "10240")

	if err := d.Run(); err != nil {
		gd.Error("Error occurs, error = %s", err.Error())
		return
	}
}
```

---
**[config]**  
So far, it only supports configuration with ini in gd. Of course, it supports more and more format configuration in future.
What's more, your configuration file must have the necessary parameters, like this:

```ini
[Log]
enable     = true
level      = "DEBUG"
logDir     = "log"

[Process]
maxCPU     = 2
maxMemory  = "2g"
healthPort = 9527

[Server]
serverName = "gd"
httpPort   = 10240
rpcPort    = 10241
```

**Process.maxCPU**: a limit of CPU usage. 0 is default, means to use half cores.  
**Process.maxMemory**: a limit of memory usage.
**Process.healthPort**: the port for monitor. If it is 0, monitor server will not run. 
**Server.serverName**: server name.  
**Server.httpPort**: http port. If it is 0, http server will not run.   
**Server.rpcPort**: rpc port. If it is 0, rpc server will not run. 

Those items mentioned above are the base need of a server application. And they are defined in config file: sample/conf/conf.json.

---
**[net]**  
provides golang network server, it is contain http server and rpc server. It is a simple demo that you can develop it on the basis of it.
I will import introduce rpc server. Focus on the rpc server.

```go
type Packet interface {
    ID() uint32
    SetErrCode(code uint32)
}

default rpc packet:
type RpcPacket struct {
    Seq       uint32
    ErrCode   uint32
    Cmd       uint32 // also be a string, for dispatch.
    PacketLen uint32
    Body      []byte
}

gd rpc packet:
type DogPacket struct {
    Header
    Body []byte
}

type Header struct {
    PacketLen uint32
    Seq       uint32
    Cmd       uint32
    CheckSum  uint32
    ErrCode   uint32
    Version   uint8
    Padding   uint8
    SOH       uint8
    EOH       uint8
}
```
The Packet is a interface in rpc server and client. So, you can make your protocol that suits yourself by implementing packet's methods, if you need.
You add new RpcPacket according to yourself rule. DogPacket is a protocol that is used by author. Of course, the author encourages the use of DogPacket. 

---
**[server]**  
provides server register and discovery. Load balancing will be provided in the future.
Service discovery registration based on etcd and zookeeper implementation.

```go
register :
    type DogRegister interface {
        NewRegister(hosts []string, root, environ, group, service string)
        SetRootNode(node string) error
        GetRootNode() (root string)
        SetHeartBeat(heartBeat time.Duration)
        SetOffline(offline bool)
        Run(ip string, port int, weight uint64) error
        Close()
    }
    
discovery :
    type DogDiscovery interface {
        NewDiscovery(dns []string)
        Watch(node string) error
        WatchMulti(nodes []string) error
        AddNode(node string, info *server.NodeInfo)
        DelNode(node string, key string)
        GetNodeInfo(node string) (nodesInfo []server.NodeInfo)
        Run() error
        Close() error
    }
    
nodeInfo:
    type NodeInfo interface {
        GetIp() string
        GetPort() int
        GetOffline() bool
        GetWeight() uint64
    }
    
    type DefaultNodeInfo struct {
        Ip      string `json:"ip"`
        Port    int    `json:"port"`
        Offline bool   `json:"offline"`
        Weight  uint64 `json:"weight"`
    }
```
The DogRegister and DogDiscovery are interface, gd supports zookeeper and etcd, so you can use others.
The NodeInfo is info of node.

---
## Usage
This example simply demonstrates the use of the gd. of course, you need to make conf.json in conf Folder. The example use service discovery with etcd. So, you can install etcd
 in your computer. Of course, you can choose to comment out these lines of code.

server:
```go
package main

import (
	"github.com/chuck1024/gd"
	de "github.com/chuck1024/gd/derror"
	"github.com/chuck1024/gd/net/dhttp"
	"github.com/chuck1024/gd/net/dogrpc"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TestReq struct {
	Data string
}

type TestResp struct {
	Ret string
}

func HandlerHttpTest(c *gin.Context, req *TestReq) (code int, message string, err error, ret *TestResp) {
	gd.Debug("httpServerTest req:%v", req)

	ret = &TestResp{
		Ret: "ok!!!",
	}

	return http.StatusOK, "ok", nil, ret
}

func HandlerRpcTest(req *TestReq) (code uint32, message string, err error, ret *TestResp) {
	gd.Debug("rpc sever req:%v", req)

	ret = &TestResp{
		Ret: "ok!!!",
	}

	return uint32(de.RpcSuccess), "ok", nil, ret
}

func Register(e *gd.Engine) {
	// http
	e.HttpServer.SetInit(func(g *gin.Engine) error {
		r := g.Group("")
		r.Use(
			dhttp.GlFilter(),
			dhttp.StatFilter(),
			dhttp.GroupFilter(),
			dhttp.Logger("sample"),
		)

		e.HttpServer.POST(r, "test", HandlerHttpTest)

		if err := e.HttpServer.CheckHandle(); err != nil {
			return err
		}

		return nil
	})

	// Rpc
	e.RpcServer.AddDogHandler(1024, HandlerRpcTest)
	if err := e.RpcServer.DogRpcRegister(); err != nil {
		gd.Error("DogRpcRegister occur error:%s", err)
		return
	}
	dogrpc.InitFilters([]dogrpc.Filter{&dogrpc.GlFilter{}, &dogrpc.LogFilter{}})
}

func main() {
	d := gd.Default()

	Register(d)

	err := d.Run()
	if err != nil {
		gd.Error("Error occurs, error = %s", err.Error())
		return
	}
}

// you can use command to test http service.
// curl -X POST http://127.0.0.1:10240/test -H "Content-Type: application/json" --data '{"Data":"test"}'
```
>* You can find it in "sample/service.go"
>* use `control+c` to stop process

rpc client:
```go
package main

import (
    "fmt"
    "github.com/chuck1024/gd"
    "github.com/chuck1024/gd/dlog"
    "github.com/chuck1024/gd/server/discovery"
    "time"
)

func main() {
    d := gd.Default()
    c := d.NewRpcClient(time.Duration(500*time.Millisecond), 0)
    // discovery 
    var r discovery.DogDiscovery
    r = &discovery.ZkDiscovery{}
    r.NewDiscovery([]string{"localhost:2379"})
    r.Watch("/root/github/gd/stagging/pool")
    r.Run()
    time.Sleep(100*time.Millisecond)
   
    hosts := r.GetNodeInfo("/root/github/gd/stagging/pool")
    for _,v := range hosts {
        dlog.Debug("%s:%d",v.GetIp(),v.GetPort())
    }
   
    // you can choose one or use load balance algorithm to choose best one.
    // or put all to c.Addr
    for _, v := range hosts {   
        if !v.GetOffline() {
            c.AddAddr(fmt.Sprintf("%s:%d", v.GetIp(), v.GetPort()))
        }
    }
    
    body := []byte("How are you?")

    code, rsp, err := c.DogInvoke(1024, body)
    if err != nil {
        dlog.Error("Error when sending request to server: %s", err)
    }

    // or use rpc protocol
    //rsp, err = c.Invoke(1024, body)
    //if err != nil {
        //t.Logf("Error when sending request to server: %s", err)
    //}

    dlog.Debug("code=%d,resp=%s", code, string(rsp))
}
```
>* It contained "sample/rpc_client.go"

---
`net module` you also use it to do something if you want to use `net module` only. Here's how it's used.

`rpc_server` show how to start rpc server
```go
package dogrpc_test

import (
	"github.com/chuck1024/gd/net/dogrpc"
	"testing"
)

func TestRpcServer(t *testing.T) {
	d := dogrpc.NewRpcServer()
	// Rpc
	d.AddHandler(1024, func(req []byte) (uint32, []byte) {
		t.Logf("rpc server request: %s", string(req))
		code := uint32(0)
		resp := []byte("Are you ok?")
		return code, resp
	})

	err := d.Run(10241)
	if err != nil {
		t.Logf("Error occurs, error = %s", err.Error())
		return
	}
}

```
>* You can find it in "net/dogrpc/rpc_server_test.go"

`rpc_client`show how to call rpc server
```go
package dogrpc_test

import (
    "github.com/chuck1024/gd"
    "github.com/chuck1024/gd/utls"
    "testing"
    "time"
)

func TestRpcClient(t *testing.T) {
    d := gd.Default()
    c := d.NewRpcClient(time.Duration(500*time.Millisecond), 0)
    c.AddAddr(utils.GetLocalIP() + ":10241")

    body := []byte("How are you?")

    code, rsp, err := c.Invoke(1024, body)
    if err != nil {
        t.Logf("Error when sending request to server: %s", err)
    }

    t.Logf("code=%d,resp=%s", code, string(rsp))
}
```
>* You can find it in "net/dogrpc/rpc_client_test.go"

---
`derror module` provides the relation usages of error. It supports the structs of CodeError which contains code, error type,
and error msg.

```go
package derror

type CodeError struct {
    errCode int
    errType string
    errMsg  string
}

var (
    RpcSuccess     = 0
    Success        = 200
    BadRequest     = 400
    Unauthorized   = 401
    Forbidden      = 403
    NotFound       = 404
    SystemError    = 500
    ParameterError = 600
    DBError        = 701
    CacheError     = 702
    RpcTimeout             = 10001
    RpcOverflow            = 10002
    RpcInternalServerError = 10003
    RpcInvalidParam        = 10004
    UnknownError = "unknown error"

    ErrMap = map[int]string{
        RpcSuccess:     "ok",
        Success:        "ok",
        BadRequest:     "bad request",
        Unauthorized:   "Unauthorized",
        Forbidden:      "Forbidden",
        NotFound:       "not found",
        SystemError:    "system error",
        ParameterError: "Parameter error",
        DBError:        "db error",
        CacheError:     "cache error",
        RpcTimeout:             "timeout error",
        RpcOverflow:            "overflow error",
        RpcInternalServerError: "interval server error",
        RpcInvalidParam:        "invalid param",
    }
)

// get derror type. you can also add type to ErrMap.
func GetErrorType(code int) string {
    t, ok := ErrMap[code]
    if !ok {
        t = UnknownError
    }
    return t
}
```

---
`server module` 
>* if you use etcd, you must download etcd module
>* `go get github.com/coreos/etcd/clientv3`
>* you can find it usage on "server/register/register_test.go" and "server/discovery/discovery.go"

```go
package register_test

import (
	"github.com/chuck1024/gd/server/register"
	"testing"
	"time"
)

func TestEtcd(t *testing.T){
    var r register.DogRegister
    r = &register.EtcdRegister{}
    r.NewRegister([]string{"localhost:2379"}, "/root/", "stagging","gd", "test", )

    r.Run("127.0.0.1", 10240,10)
    time.Sleep(3 * time.Second)
    r.Close()
}

func TestZk(t *testing.T){
    var r register.DogRegister
    r = &register.ZkRegister{}
    r.NewRegister([]string{"localhost:2181"}, "/root/", "stagging","gd", "test", )
    r.Run("127.0.0.1", 10240,10)
    time.Sleep(10 * time.Second)
    r.Close()
}
```
```go
package discovery_test

import (
    "github.com/chuck1024/gd/server/discovery"
    "testing"
    "time"
)

func TestDiscEtcd(t *testing.T){
    var r discovery.DogDiscovery
    r = &discovery.EtcdDiscovery{}
    r.NewDiscovery([]string{"localhost:2379"})
    r.Watch("/root/gd/test/stagging/pool")
    r.Run()
    time.Sleep(100*time.Millisecond)

    n1 := r.GetNodeInfo("/root/gd/test/stagging/pool")
    for _,v := range n1 {
        t.Logf("%s:%d",v.GetIp(),v.GetPort())
    }

    time.Sleep(10*time.Second)
}

func TestDiscZk(t *testing.T){
    var r discovery.DogDiscovery
    r = &discovery.ZkDiscovery{}
    r.NewDiscovery([]string{"localhost:2181"})
    r.Watch("/root/gd/test/stagging/pool")
    r.Run()
    time.Sleep(100*time.Millisecond)
    n1 := r.GetNodeInfo("/root/gd/test/stagging/pool")
    for _,v := range n1 {
        t.Logf("%s:%d",v.GetIp(),v.GetPort())
    }
    time.Sleep(10*time.Second)
}
```

More information can be obtained in the source code

---
## License

gd is released under the [**MIT LICENSE**](http://opensource.org/licenses/mit-license.php).  
