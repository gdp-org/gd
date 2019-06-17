# GoDog

"go" is the meaning of a dog in Chinese pronunciation, and dog's original intention is also a dog. So godog means "狗狗" in Chinese, which is very cute.

---
## Author

```
author: Chuck1024
email : chuck.ch1024@outlook.com
```

---
## Installation

Start with cloning godog[2.0]:

```
> go get github.com/chuck1024/godog
```

---
## Introduction

GoDog is a basic framework implemented by golang, which is aiming at helping developers setup feature-rich server quickly.

The framework contains `config module`,`error module`,`net module` and `server module`. You can select any modules according to your practice. More features will be added later. I hope anyone who is interested in this work can join it and let's enhance the system function of this framework together.

>* [gin](https://github.com/gin-gonic/gin) [etcd](https://github.com/etcd-io/etcd) and [zookeeper](https://github.com/samuel/go-zookeeper) are third-party library. 
>* Authors are [**Gin-Gonic**](https://gin-gonic.com/),[**etcd-io**](https://github.com/etcd-io) and [**Samuel Stauffer**](https://github.com/samuel).Thanks for them here. 

---
## Quick start
```go
package main

import (
    "github.com/bitly/go-simplejson"
    "github.com/chuck1024/doglog"
    "github.com/chuck1024/godog"
    "github.com/gin-gonic/gin"
    "net/http"
)

func HandlerHttpTest(c *gin.Context, req *simplejson.Json) (code int, message string, err error, ret string) {
    doglog.Debug("httpServerTest req:%v", req)
    
    ret = "ok!!!"
    return http.StatusOK, "ok", nil, ret
}

func main() {
    d := godog.Default()
    d.HttpServer.DefaultAddHandler("test", HandlerHttpTest)
    d.HttpServer.DefaultRegister()
    
    d.Config.BaseConfig.Server.HttpPort = 10240
    err := d.Run()
    if err != nil {
        doglog.Error("Error occurs, error = %s", err.Error())
        return
    }
}
```

---
**[config]**  
So far, it only supports configuration with json in godog. Of course, it supports more and more format configuration in future.
What's more, your configuration file must have the necessary parameters, like this:
```json
{
  "Log": "conf/log.xml",
  "Prog": {
    "CPU": 0,
    "HealthPort": 0
  },
  "Server": {
    "AppName": "godog",
    "HttpPort": 10240,
    "TcpPort": 10241
  }
}
```

**Prog.CPU**: a limit of CPU usage. 0 is default, means to use all cores.  
**Prog.HealthPort**: the port for monitor. If it is 0, monitor server will not run. 
**Server.AppName**: server name.  
**Server.HttpPort**: http port. If it is 0, http server will not run.   
**Server.TcpPort**: tcp port. If it is 0, tcp server will not run. 

Those items mentioned above are the base need of a server application. And they are defined in config file: sample/conf/conf.json.

---
**[net]**  
provides golang network server, it is contain http server and tcp server. It is a simple demo that you can develop it on the basis of it.
I will import introduce tcp server. Focus on the tcp server.

```go
type Packet interface {
    ID() uint32
    SetErrCode(code uint32)
}

default tcp packet:
type TcpPacket struct {
    Seq       uint32
    ErrCode   uint32
    Cmd       uint32 // also be a string, for dispatch.
    PacketLen uint32
    Body      []byte
}

godog tcp packet:
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
The Packet is a interface in tcp server and client. So, you can make your protocol that suits yourself by implementing packet's methods, if you need.
You add new TcpPacket according to yourself rule. DogPacket is a protocol that is used by author. Of course, the author encourages the use of DogPacket. 

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
The DogRegister and DogDiscovery are interface, godog supports zookeeper and etcd, so you can use others.
The NodeInfo is info of node.

---
## Usage
This example simply demonstrates the use of the godog. of course, you need to make conf.json in conf Folder. The example use service discovery with etcd. So, you can install etcd
 in your computer. Of course, you can choose to comment out these lines of code.

service:
```go
package main

import (
	"github.com/chuck1024/doglog"
	"github.com/chuck1024/godog"
	"github.com/chuck1024/godog/net/httplib"
	"github.com/chuck1024/godog/server/register"
	"github.com/chuck1024/godog/utils"
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
	doglog.Debug("httpServerTest req:%v", req)

	ret = &TestResp{
		Ret: "ok!!!",
	}

	return http.StatusOK, "ok", nil, ret
}

func HandlerTcpTest(req []byte) (uint32, []byte) {
	doglog.Debug("tcp server request: %s", string(req))
	code := uint32(200)
	resp := []byte("Are you ok?")
	return code, resp
}

func main() {
	d := godog.Default()
	// Http
	d.HttpServer.DefaultAddHandler("test", HandlerHttpTest)
	d.HttpServer.DefaultRegister()

	// default tcp server, you can choose godog tcp server
	//d.TcpServer = tcplib.NewDogTcpServer()

	// Tcp
	d.TcpServer.AddTcpHandler(1024, HandlerTcpTest)

	// register params
	etcdHost, _ := d.Config.Strings("etcdHost")
	root, _ := d.Config.String("root")
	environ, _ := d.Config.String("environ")
	group, _ := d.Config.String("group")
	weight, _ := d.Config.Int("weight")

	// register
	var r register.DogRegister
	r = &register.EtcdRegister{}
	r.NewRegister(etcdHost, root, environ, group, d.Config.BaseConfig.Server.AppName)
	r.Run(utils.GetLocalIP(), d.Config.BaseConfig.Server.TcpPort, uint64(weight))

	err := d.Run()
	if err != nil {
		doglog.Error("Error occurs, error = %s", err.Error())
		return
	}
}

// you can use command to test http service.
// curl http://127.0.0.1:10240/test

```
>* You can find it in "sample/service.go"
>* use `control+c` to stop process

tcp client:
```go
package main

import (
    "fmt"
    "github.com/chuck1024/doglog"
    "github.com/chuck1024/godog"
    "github.com/chuck1024/godog/server/discovery"
    "time"
)

func main() {
    d := godog.Default()
    c := d.NewTcpClient(500, 0)
    // discovery 
    var r discovery.DogDiscovery
    r = &discovery.EtcdDiscovery{}
    r.NewDiscovery([]string{"localhost:2379"})
    r.Watch("/root/github/godog/stagging/pool")
    r.Run()
    time.Sleep(100*time.Millisecond)
   
    hosts := r.GetNodeInfo("/root/github/godog/stagging/pool")
    for _,v := range hosts {
        doglog.Debug("%s:%d",v.GetIp(),v.GetPort())
    }
   
    // you can choose one or use load balance algorithm to choose best one.
    // or put all to c.Addr
    for _, v := range hosts {   
        if !v.GetOffline() {
            c.AddAddr(fmt.Sprintf("%s:%d", v.GetIp(), v.GetPort()))
        }
    }
    
    body := []byte("How are you?")

    rsp, err := c.Invoke(1024, body)
    if err != nil {
        doglog.Error("Error when sending request to server: %s", err)
    }

    // or use godog protocol
    //rsp, err = c.DogInvoke(1024, body)
    //if err != nil {
        //t.Logf("Error when sending request to server: %s", err)
    //}

    doglog.Debug("resp=%s", string(rsp))
}
```
>* It contained "sample/tcp_client.go"

---
`net module` you also use it to do something if you want to use `net module` only. Here's how it's used.

`tcp_server` show how to start tcp server
```go
package tcplib_test

import (
	"github.com/chuck1024/godog"
	"testing"
)

func TestTcpServer(t *testing.T) {
	d := godog.Default()
	// Tcp
	d.TcpServer.AddTcpHandler(1024, func(req []byte) (uint32, []byte) {
		t.Logf("tcp server request: %s", string(req))
		code := uint32(0)
		resp := []byte("Are you ok?")
		return code, resp
	})

	err := d.TcpServer.Run(10241)
	if err != nil {
		t.Logf("Error occurs, error = %s", err.Error())
		return
	}
}

```
>* You can find it in "net/tcplib/tcp_server_test.go"

`tcp_client`show how to call tcp server
```go
package tcplib_test

import (
    "github.com/chuck1024/godog"
    "github.com/chuck1024/godog/utils"
    "testing"
)

func TestTcpClient(t *testing.T) {
    d := godog.Default()
    c := d.NewTcpClient(500, 0)
    c.AddAddr(utils.GetLocalIP() + ":10241")

    body := []byte("How are you?")

    rsp, err := c.Invoke(1024, body)
    if err != nil {
        t.Logf("Error when sending request to server: %s", err)
    }

    t.Logf("resp=%s", string(rsp))
}
```
>* You can find it in "net/tcplib/tcp_client_test.go"

---
`config module` provides the related configuration of the project.
>* You can find it in "sample/config_test.go"

```go
package main_test

import (
	"github.com/chuck1024/doglog"
    "github.com/chuck1024/godog"
    "testing"
)

func TestConfig(t *testing.T) {
	// init log
	doglog.LoadConfiguration("conf/log.xml")

    // Notice: config contains BaseConfigure. config.json must contain the BaseConfigure configuration.
    // The location of config.json is "conf/conf.json". Of course, you change it if you want.

    d := godog.Default()
    // AppConfig.BaseConfig.Server.AppName is service name
    name := d.Config.BaseConfig.Server.AppName
    t.Logf("name:%s", name)

    // you can add configuration items directly in conf.json
    stringValue, err := d.Config.String("stringKey")
    if err != nil {
        doglog.Error("get key occur error: %s", err)
        return
    }
    t.Logf("value:%s", stringValue)

    stringsValue, err := d.Config.Strings("stringsKey")
    if err != nil {
        doglog.Error("get key occur error: %s", err)
        return
    }
    t.Logf("value:%s", stringsValue)
    
    intValue, err := d.Config.Int("intKey")
    if err != nil {
        doglog.Error("get key occur error: %s", err)
        return
    }
    t.Logf("value:%d", intValue)

    BoolValue, err := d.Config.Bool("boolKey")
    if err != nil {
        doglog.Error("get key occur error: %s", err)
        return
    }
    t.Logf("value:%t", BoolValue)

    // you can add config key-value if you need.
    d.Config.Set("yourKey", "yourValue")

    // get config key
    yourValue, err := d.Config.String("yourKey")
    if err != nil {
        doglog.Error("get key occur error: %s", err)
        return
    }
    t.Logf("yourValue:%s", yourValue)
}
```

---
`error module` provides the relation usages of error. It supports the structs of CodeError which contains code, error type,
and error msg.

```go
package error

type CodeError struct {
    errCode int
    errType string
    errMsg  string
}

var (
    TcpSuccess     = 0
    Success        = 200
    BadRequest     = 400
    Unauthorized   = 401
    Forbidden      = 403
    NotFound       = 404
    SystemError    = 500
    ParameterError = 600
    DBError        = 701
    CacheError     = 702
    UnknownError = "unknown error"

    ErrMap = map[int]string{
        TcpSuccess:     "ok",
        Success:        "ok",
        BadRequest:     "bad request",
        Unauthorized:   "Unauthorized",
        Forbidden:      "Forbidden",
        NotFound:       "not found",
        SystemError:    "system error",
        ParameterError: "Parameter error",
        DBError:        "db error",
        CacheError:     "cache error",
    }
)

// get error type. you can also add type to ErrMap.
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
	"github.com/chuck1024/godog/server/register"
	"testing"
	"time"
)

func TestEtcd(t *testing.T){
    var r register.DogRegister
    r = &register.EtcdRegister{}
    r.NewRegister([]string{"localhost:2379"}, "/root/", "stagging","godog", "test", )

    r.Run("127.0.0.1", 10240,10)
    time.Sleep(3 * time.Second)
    r.Close()
}

func TestZk(t *testing.T){
    var r register.DogRegister
    r = &register.ZkRegister{}
    r.NewRegister([]string{"localhost:2181"}, "/root/", "stagging","godog", "test", )
    r.Run("127.0.0.1", 10240,10)
    time.Sleep(10 * time.Second)
    r.Close()
}
```
```go
package discovery_test

import (
    "github.com/chuck1024/godog/server/discovery"
    "testing"
    "time"
)

func TestDiscEtcd(t *testing.T){
    var r discovery.DogDiscovery
    r = &discovery.EtcdDiscovery{}
    r.NewDiscovery([]string{"localhost:2379"})
    r.Watch("/root/godog/test/stagging/pool")
    r.Run()
    time.Sleep(100*time.Millisecond)

    n1 := r.GetNodeInfo("/root/godog/test/stagging/pool")
    for _,v := range n1 {
        t.Logf("%s:%d",v.GetIp(),v.GetPort())
    }

    time.Sleep(10*time.Second)
}

func TestDiscZk(t *testing.T){
    var r discovery.DogDiscovery
    r = &discovery.ZkDiscovery{}
    r.NewDiscovery([]string{"localhost:2181"})
    r.Watch("/root/godog/test/stagging/pool")
    r.Run()
    time.Sleep(100*time.Millisecond)
    n1 := r.GetNodeInfo("/root/godog/test/stagging/pool")
    for _,v := range n1 {
        t.Logf("%s:%d",v.GetIp(),v.GetPort())
    }
    time.Sleep(10*time.Second)
}
```

More information can be obtained in the source code

---
## License

godog is released under the [**MIT LICENSE**](http://opensource.org/licenses/mit-license.php).  
