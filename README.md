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

The framework contains `config module`,`error module`,`log module`,`net module` and `dao module`. You can select any modules according to your practice. More features will be added later. I hope anyone who is interested in this work can join it and let's enhance the system function of this framework together.

>* [logging](https://github.com/xuyu/logging),[redigo](https://github.com/garyburd/redigo/redis) and [redis-go-cluster](https://github.com/chasex/redis-go-cluster) are third-party library. 
>* Authors are [**xuyu**](https://github.com/xuyu),[**garyburd**](https://github.com/garyburd) and [**chasex**](https://github.com/chasex).Thanks for them here. 
>* I modified the `logging module`, adding the printing of file name, row number and time.

## Usage

`net module` provides golang network server, it is contain http server and tcp server. It is a simple demo that you can develop it on the basis of it.

Focus on the tcp server.
```
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

>* You can find it in "test/service.go"
>* use `control+c` to stop process

```
func HandlerHttpTest(w http.ResponseWriter, r *http.Request) {
    godog.Debug("connected : %s", r.RemoteAddr)
    w.Write([]byte("test success!!!"))
}

func HandlerTcpTest(req []byte) (uint16, []byte) {
    godog.Debug("tcp server request: %s", string(req))
    code := uint16(200)
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

    err := godog.Run()
    if err != nil {
        godog.Error("Error occurs, error = %s", err.Error())
        return
    }
}

// you can use command to test service that it is in another file <serviceTest.txt>.
```

`tcpclient` show how to call tcp server
>* You can find it in "test/tcp_client_test.go"

```
func TestTcpClient(t *testing.T) {
    c := godog.NewTcpClient(500, 0)
    // remember alter addr
    c.AddAddr("127.0.0.1:10241")

    body := []byte("How are you?")

    rsp, err := c.Invoke(1024, body)
    if err != nil {
        godog.Error("Error when sending request to server: %s", err)
    }

    // or use godog protocol
    //rsp, err = c.DogInvoke(1024, body)
    //if err != nil {
        //godog.Error("Error when sending request to server: %s", err)
    //}

    godog.Debug("resp=%s", string(rsp))
}

```

`config module` provides the related configuration of the project.
>* You can find it in "test/config_test.go"

```
func TestConfig(t *testing.T) {
    // Notice: config contains BaseConfigure. config.json must contain the BaseConfigure configuration.
    // The location of config.json is "conf/conf.json". Of course, you change it if you want.

    // AppConfig.BaseConfig.Log.File is the path of log file.
    file := godog.AppConfig.BaseConfig.Log.File
    godog.Debug("log file:%s", file)

    // AppConfig.BaseConfig.Log.Level is log level.
    // DEBUG   logLevel = 1
    // INFO    logLevel = 2
    // WARNING logLevel = 3
    // ERROR   logLevel = 4
    // DISABLE logLevel = 255
    level := godog.AppConfig.BaseConfig.Log.Level
    godog.Debug("log level:%s", level)

    // AppConfig.BaseConfig.Server.AppName is service name
    name := godog.AppConfig.BaseConfig.Server.AppName
    godog.Debug("name:%s", name)

    // AppConfig.BaseConfig.Log.Suffix is suffix of log file.
    // suffix = "060102-15" . It indicates that the log is cut per hour
    // suffix = "060102" . It indicates that the log is cut per day
    suffix := godog.AppConfig.BaseConfig.Log.Suffix
    godog.Debug("log suffix:%s", suffix)

    // you can add configuration items directly in conf.json
    stringValue, err := godog.AppConfig.String("stringKey")
    if err != nil {
        godog.Error("get key occur error: %s", err)
        return
    }
    godog.Debug("value:%s", stringValue)

    intValue, err := godog.AppConfig.Int("intKey")
    if err != nil {
        godog.Error("get key occur error: %s", err)
        return
    }
    godog.Debug("value:%d", intValue)

    BoolValue, err := godog.AppConfig.Bool("boolKey")
    if err != nil {
        godog.Error("get key occur error: %s", err)
        return
    }
    godog.Debug("value:%t", BoolValue)

    // you can add config key-value if you need.
    godog.AppConfig.Set("yourKey", "yourValue")

    // get config key
    yourValue, err := godog.AppConfig.String("yourKey")
    if err != nil {
        godog.Error("get key occur error: %s", err)
        return
    }
    godog.Debug("yourValue:%s", yourValue)
}

```

`error module` provides the relation usages of error that you can find it in godog.

`dao module` provides the relation usages of db and redis.
>* You can find it in "test/db_test.go" and "test/redis_test.go"

```
func TestAdd(t *testing.T) {
    url, err := godog.AppConfig.String("mysql")
    if err != nil {
        godog.Warning("[init] get config mysql url occur error: ", err)
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
        godog.Error("[testAdd] errors occur while res.RowsAffected(): %s", err.Error())
        return
    }
}

func TestUpdate(t *testing.T) {
    url, err := godog.AppConfig.String("mysql")
    if err != nil {
        godog.Warning("[init] get config mysql url occur error: ", err)
        return
    }

    MysqlHandle = db.Init(url)
    
    td := &TestDB{
        CardId: 1024,
    }

    if err := td.Update(1025); err != nil {
        godog.Error("[testUpdate] errors occur while res.RowsAffected(): %s", err.Error())
        return
    }
}

func TestQuery(t *testing.T) {
    url, err := godog.AppConfig.String("mysql")
    if err != nil {
        godog.Warning("[init] get config mysql url occur error: ", err)
        return
    }

    MysqlHandle = db.Init(url)
    
    td := &TestDB{}

    tt, err := td.Query(1024)
    if err != nil {
        godog.Error("query occur error:", err)
        return
    }

    godog.Debug("query: %v", *tt)
}
```

```
func TestRedis(t *testing.T) {
    URL,_ := godog.AppConfig.String("redis")
    cache.Init(URL)
    
    key := "key"
    err := cache.Set( key, "value")
    if err != nil {
        godog.Error("redis set occur error:%s", err)
        return
    }

    godog.Debug("set success:%s",key)

    value, err := cache.Get(key)
    if err != nil {
        godog.Error("redis get occur error: %s", err)
        return
    }
    godog.Debug("get value: %s",value)
}
```
## License

godog is released under the [**MIT LICENSE**](http://opensource.org/licenses/mit-license.php).  
