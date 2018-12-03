# GoDog

"go" is the meaning of a dog in Chinese pronunciation, and dog's original intention is also a dog. So godog means "狗狗" in Chinese, which is very cute.


## Author

```
author: Chuck1024
email : chuck.ch1024@outlook.com
```

## Installation

Start with cloning godog:

```
> go get github.com/chuck1024/godog
```

## Introduction

GoDog is a basic framework implemented by golang, which is aiming at helping developers setup feature-rich server quickly.

The framework contains `config module`,`error module`,`log module`,`net module`,`server module` and `dao module`. You can select any modules according to your practice. More features will be added later. I hope anyone who is interested in this work can join it and let's enhance the system function of this framework together.

>* [logging](https://github.com/xuyu/logging),[redigo](https://github.com/garyburd/redigo/redis) and [redis-go-cluster](https://github.com/chasex/redis-go-cluster) are third-party library. 
>* Authors are [**xuyu**](https://github.com/xuyu),[**garyburd**](https://github.com/garyburd) and [**chasex**](https://github.com/chasex).Thanks for them here. 
>* I modified the `logging module`, adding the printing of file name, row number and time.

## Usage
This example simply demonstrates the use of the godog. of course, you need to make conf.json in conf Folder. The example use service discovery with etcd. So, you can install etcd
 in your computer. Of course, you can choose to comment out these lines of code.
```go
package main

import (
    "github.com/chuck1024/godog"
    "github.com/chuck1024/godog/server/register"
    "github.com/chuck1024/godog/utils"
    "net/http"
)

func HandlerHttpTest(w http.ResponseWriter, r *http.Request) {
    godog.Debug("connected : %s", r.RemoteAddr)
    w.Write([]byte("test success!!!"))
}

func HandlerTcpTest(req []byte) (uint32, []byte) {
    godog.Debug("tcp server request: %s", string(req))
    code := uint32(200)
    resp := []byte("Are you ok?")
    return code, resp
}

func main() {
    // Http
    godog.AppHttp.AddHttpHandler("/test", HandlerHttpTest)

    // default tcp server, you can choose godog tcp server
    //godog.AppTcp = tcplib.AppDog

    // Tcp
    godog.AppTcp.AddTcpHandler(1024, HandlerTcpTest)

    // register params
    etcdHost, _ := godog.AppConfig.Strings("etcdHost")
    root, _ := godog.AppConfig.String("root")
    environ, _ := godog.AppConfig.String("environ")
    group, _ := godog.AppConfig.String("group")
    weight, _ := godog.AppConfig.Int("weight")
    
    // register
    var r register.DogRegister
    r = &register.EtcdRegister{}
    r.NewRegister(etcdHost, root, environ, group, godog.AppConfig.BaseConfig.Server.AppName, )
    r.Run(utils.GetLocalIP(), godog.AppConfig.BaseConfig.Server.TcpPort, uint64(weight))
    
    err := godog.Run()
    if err != nil {
        godog.Error("Error occurs, error = %s", err.Error())
        return
    }
}
// you can use command to test http service.
// curl http://127.0.0.1:10240/test
```
>* You can find it in "example/service.go"
>* use `control+c` to stop process

```go
package main

import (
    "fmt"
    "github.com/chuck1024/godog"
    "github.com/chuck1024/godog/server/discovery"
    "time"
)

func main() {
    c := godog.NewTcpClient(500, 0)
    // remember alter addr
    var r discovery.DogDiscovery
    r = &discovery.EtcdDiscovery{}
    r.NewDiscovery([]string{"localhost:2379"})
    r.Watch("/root/github/godog/stagging/pool")
    r.Run()
    time.Sleep(100*time.Millisecond)
   
    hosts := r.GetNodeInfo("/root/github/godog/stagging/pool")
    for _,v := range hosts {
        godog.Debug("%s:%d",v.GetIp(),v.GetPort())
    }
   
    // you can choose one
    c.AddAddr(hosts[0].GetIp() + ":" + fmt.Sprintf("%d",hosts[0].GetPort()))

    body := []byte("How are you?")

    rsp, err := c.Invoke(1024, body)
    if err != nil {
        godog.Error("Error when sending request to server: %s", err)
    }

    // or use godog protocol
    //rsp, err = c.DogInvoke(1024, body)
    //if err != nil {
        //t.Logf("Error when sending request to server: %s", err)
    //}

    godog.Debug("resp=%s", string(rsp))
}
```
>* It contained "example/tcp_client.go"


`net module` provides golang network server, it is contain http server and tcp server. It is a simple demo that you can develop it on the basis of it.

Focus on the tcp server.
```go
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

PS: of course, you can add new TcpPacket according to yourself rule.
    DogPacket is a simple. You can consult god_server.go and dog_client.go and make your own protocol.
```

`tcpserver` show how to start tcp server
```go
package tcplib_test

import (
    "github.com/chuck1024/godog/net/tcplib"
    "testing"
)

func TestTcpServer(t *testing.T) {
    // Tcp 
    tcplib.AppTcp.AddTcpHandler(1024, func(req []byte) (uint32, []byte) {
        t.Logf("tcp server request: %s", string(req))
        code := uint32(0)
        resp := []byte("Are you ok?")
        return code, resp
    })

    err := tcplib.AppTcp.Run(10241)
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
    "testing"
)

func TestTcpClient(t *testing.T) {
    c := godog.NewTcpClient(500, 0)
    c.AddAddr("127.0.0.1:10241")

    body := []byte("How are you?")

    rsp, err := c.Invoke(1024, body)
    if err != nil {
        t.Logf("Error when sending request to server: %s", err)
    }

    t.Logf("resp=%s", string(rsp))
}
```
>* You can find it in "net/tcplib/tcp_client_test.go"

`config module` provides the related configuration of the project.
>* You can find it in "example/config_test.go"

```go
package main_test

import (
    "github.com/chuck1024/godog"
    "github.com/chuck1024/godog/log"
    "testing"
)

func TestConfig(t *testing.T) {
    // init log
    log.InitLog(godog.AppConfig.BaseConfig.Log.File, godog.AppConfig.BaseConfig.Log.Level, godog.AppConfig.BaseConfig.Server.AppName, godog.AppConfig.BaseConfig.Log.Suffix, godog.AppConfig.BaseConfig.Log.Daemon)

    // Notice: config contains BaseConfigure. config.json must contain the BaseConfigure configuration.
    // The location of config.json is "conf/conf.json". Of course, you change it if you want.

    // AppConfig.BaseConfig.Log.File is the path of log file.
    file := godog.AppConfig.BaseConfig.Log.File
    t.Logf("log file:%s", file)

    // AppConfig.BaseConfig.Log.Level is log level.
    // DEBUG   logLevel = 1
    // INFO    logLevel = 2
    // WARNING logLevel = 3
    // ERROR   logLevel = 4
    // DISABLE logLevel = 255
    level := godog.AppConfig.BaseConfig.Log.Level
    t.Logf("log level:%s", level)

    // AppConfig.BaseConfig.Server.AppName is service name
    name := godog.AppConfig.BaseConfig.Server.AppName
    t.Logf("name:%s", name)

    // AppConfig.BaseConfig.Log.Suffix is suffix of log file.
    // suffix = "060102-15" . It indicates that the log is cut per hour
    // suffix = "060102" . It indicates that the log is cut per day
    suffix := godog.AppConfig.BaseConfig.Log.Suffix
    t.Logf("log suffix:%s", suffix)

    // you can add configuration items directly in conf.json
    stringValue, err := godog.AppConfig.String("stringKey")
    if err != nil {
        godog.Error("get key occur error: %s", err)
        return
    }
    t.Logf("value:%s", stringValue)

    stringsValue, err := godog.AppConfig.Strings("stringsKey")
    if err != nil {
        godog.Error("get key occur error: %s", err)
        return
    }
    t.Logf("value:%s", stringsValue)
    
    intValue, err := godog.AppConfig.Int("intKey")
    if err != nil {
        godog.Error("get key occur error: %s", err)
        return
    }
    t.Logf("value:%d", intValue)

    BoolValue, err := godog.AppConfig.Bool("boolKey")
    if err != nil {
        godog.Error("get key occur error: %s", err)
        return
    }
    t.Logf("value:%t", BoolValue)

    // you can add config key-value if you need.
    godog.AppConfig.Set("yourKey", "yourValue")

    // get config key
    yourValue, err := godog.AppConfig.String("yourKey")
    if err != nil {
        godog.Error("get key occur error: %s", err)
        return
    }
    t.Logf("yourValue:%s", yourValue)
}
```

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

`server module` provides server register and discovery. Load balancing will be provided in the future.
Service discovery registration based on etcd and zookeeper implementation.
>* if you use etcd, you must download etcd module
>* `go get github.com/coreos/etcd/clientv3`

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

`dao module` provides the relation usages of db and redis.
>* You can find it in "dao/db/db_test.go" and "dao/cache/redis_test.go"

```go
func TestAdd(t *testing.T) {
    url, err := godog.AppConfig.String("mysql")
    if err != nil {
        t.Logf("[init] get config mysql url occur error: %s", err)
        return
    }

    MysqlHandle = db.Init(url)
    
    td := &TestDB{
        Name:     "chuck",
        CardId:   1025,
        Sex:      "male",
        Birthday: 1024,
        Status:   1,
        CreateTs: uint64(time.Now().Unix()),
    }

    if err := td.Add(); err != nil {
        t.Logf("[testAdd] errors occur while res.RowsAffected(): %s", err.Error())
        return
    }
}

func TestUpdate(t *testing.T) {
    url, err := godog.AppConfig.String("mysql")
    if err != nil {
        t.Logf("[init] get config mysql url occur error:%s ", err)
        return
    }

    MysqlHandle = db.Init(url)
    
    td := &TestDB{
        CardId: 1024,
    }

    if err := td.Update(1025); err != nil {
        t.Logf("[testUpdate] errors occur while res.RowsAffected(): %s", err.Error())
        return
    }
}

func TestQuery(t *testing.T) {
    url, err := godog.AppConfig.String("mysql")
    if err != nil {
        t.Logf("[init] get config mysql url occur error: %s", err)
        return
    }

   MysqlHandle = db.Init(url)
    
    td := &TestDB{}

    tt, err := td.Query(1024)
    if err != nil {
        t.Logf("query occur error:%s", err)
        return
    }

    t.Logf("query: %v", *tt)
}
```

```go
package cache_test

import (
    "github.com/chuck1024/godog"
    "github.com/chuck1024/godog/dao/cache"
    "testing"
)

func TestRedis(t *testing.T) {
    URL,_ := godog.AppConfig.String("redis")
    cache.Init(URL)
    
    key := "key"
    err := cache.Set( key, "value")
    if err != nil {
        t.Logf("redis set occur error:%s", err)
        return
    }

    t.Logf("set success:%s",key)

    value, err := cache.Get(key)
    if err != nil {
        t.Logf("redis get occur error: %s", err)
        return
    }
    t.Logf("get value: %s",value)
}
```

More information can be obtained in the source code
## License

godog is released under the [**MIT LICENSE**](http://opensource.org/licenses/mit-license.php).  
