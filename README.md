# godog

"go" is the meaning of a dog in Chinese pronunciation, and dog's original intention is also a dog. So godog means "狗狗" in Chinese, which is very cute.


## Author

```
author: Chuck1024
email : chuck.ch1024@outlook.com
```

## Installation

Start with cloning godog:

```
> git clone https://github.com/chuck1024/godog.git
```

## Introduction

Godog is a basic framework implemented by golang, which is aiming at helping developers setup feature-rich server quickly.

The framework contains `config module`,`error module`,`logging module`,`net module`,`store mudle` and `service module`. You can select any modules according to your practice. More features will be added later. I hope anyone who is interested in this work can join it and let's enhance the system function of this framework together.

>* [logging](https://github.com/xuyu/logging)  module and [goredis](https://github.com/xuyu/goredis) module are third-party library. Author is [**xuyu**](https://github.com/xuyu). Thanks for xuyu here.
>* I modified the `log module`, adding the printing of file name, row number and time.   

## Usage

`service module` provides golang server. It is a simple demo that you can develop it on the basis of it. 
>* You can find it in "godog/test/serviceTest.go"
>* use `control+c` to stop process

```
/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"godog"
	"net/http"
)

var App *godog.Application

func HandlerHttpTest(w http.ResponseWriter, r *http.Request) {
	godog.Debug("connected : %s", r.RemoteAddr)
	w.Write([]byte("test success!!!"))
}

func HandlerTcpTest(req []byte) (uint16, []byte) {
	godog.Debug("tcp server request: %s", string(req))
	code := uint16(0)
	resp := []byte("Are you ok?")
	return code, resp
}

func main() {
	AppName := "test"
	App = godog.NewApplication(AppName)
	// Http
	App.AppHttp.AddHandlerFunc("/test", HandlerHttpTest)

	// Tcp
	App.AppTcpServer.AddTcpHandler(1024, HandlerTcpTest)

	err := App.Run()
	if err != nil {
		godog.Error("Error occurs, error = %s", err.Error())
		return
	}
}

// you can use command to test service that it is in another file <serviceTest.txt>.
```
`tcpClient` show how to call tcpserver
>* You can find it in "godog/test/tcpClientTest.go"

```
/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"godog"
	"godog/net/tcplib"
)

func main() {
	c := tcplib.NewClient(500, 0)
	// remember alter addr
	c.AddAddr("127.0.0.1:10241")

	body := []byte("How are you?")

	rsp, err := c.Invoke(1024, body)
	if err != nil {
		godog.Error("Error when sending request to server: %s", err)
	}

	godog.Debug("resp=%s", string(rsp))
}

```

`config module` provides the related configuration of the project.
>* You can find it in "godog/test/configTest.go"

```
/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/xuyu/logging" // import logging module
	"godog"
	"godog/config"
	_ "godog/log" // init log
)

var AppConfig *config.DogAppConfig

func main() {
	AppConfig = config.AppConfig

	// Notice: config contains BaseConfigure. config.json must contain the BaseConfigure configuration.
	// The location of config.json is "conf/conf.json". Of course, you change it if you want.

	// AppConfig.BaseConfig.Log.File is the path of log file.
	file := AppConfig.BaseConfig.Log.File
	godog.Debug("log file:%s", file)

	// AppConfig.BaseConfig.Log.Level is log level.
	// DEBUG   logLevel = 1
	// INFO    logLevel = 2
	// WARNING logLevel = 3
	// ERROR   logLevel = 4
	// DISABLE logLevel = 255
	level := AppConfig.BaseConfig.Log.Level
	godog.Debug("log level:%s", level)

	// AppConfig.BaseConfig.Log.Name is service name
	name := AppConfig.BaseConfig.Log.Name
	godog.Debug("name:%s", name)

	// AppConfig.BaseConfig.Log.Suffix is suffix of log file.
	// suffix = "060102-15" . It indicates that the log is cut per hour
	// suffix = "060102" . It indicates that the log is cut per day
	suffix := AppConfig.BaseConfig.Log.Suffix
	godog.Debug("log suffix:%s", suffix)

	// you can add configuration items directly in conf.json
	stringValue, err := AppConfig.String("stringKey")
	if err != nil {
		logging.Error("get key occur error: %s", err)
		return
	}
	godog.Debug("value:%s", stringValue)

	intValue, err := AppConfig.Int("intKey")
	if err != nil {
		logging.Error("get key occur error: %s", err)
		return
	}
	godog.Debug("value:%d", intValue)

	BoolValue, err := AppConfig.Bool("boolKey")
	if err != nil {
		logging.Error("get key occur error: %s", err)
		return
	}
	godog.Debug("value:%t", BoolValue)

	// you can add config key-value if you need.
	AppConfig.Set("yourKey", "yourValue")

	// get config key
	yourValue, err := AppConfig.String("yourKey")
	if err != nil {
		logging.Error("get key occur error: %s", err)
		return
	}
	godog.Debug("yourValue:%s", yourValue)
}
```

`error module` provides the relation usages of error that you can find it in godog.

`store module` provides the relation usages of db and redis.
>* You can find it in "godog/test/dbTest.go" and "godog/test/redisTest.go"

```
/**
 * Copyright 2018 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"errors"
	"fmt"
	"godog"
	"godog/store/db"
	"time"
)

const (
	tableName = "test"
)

type Test struct {
	Name     string `json:"name"`
	CardId   uint64 `json:"card_id"`
	Sex      string `json:"sex"`
	Birthday uint64 `json:"birthday"`
	Status   uint8  `json:"status"`
	CreateTs uint64 `json:"create_time"`
}

func (t *Test) Add() error {
	insertData := map[string]interface{}{
		"name":        t.Name,
		"card_id":     t.CardId,
		"sex":         t.Sex,
		"birthday":    t.Birthday,
		"status":      t.Status,
		"create_time": t.CreateTs,
	}

	sql := db.InsertOne(tableName, insertData)
	stmt, err := db.MysqlHandle.Prepare(sql)
	if err != nil {
		godog.Error("errors occur while util.Db_zone.Prepare(): %s", err.Error())
		return err
	}

	defer stmt.Close()

	res, err := stmt.Exec(t.Name, t.CardId, t.Sex, t.Birthday, t.Status, t.CreateTs)
	if err != nil {
		godog.Error("errors occur while stmt.Exec(): %s", err.Error())
		return err
	}

	num, err := res.RowsAffected()
	if err != nil {
		godog.Error("errors occur while res.RowsAffected(): %s", err.Error())
		return err
	}

	if num != 1 {
		return errors.New("none row Affected")
	}

	return nil
}

func (t *Test) Update(birthday uint64) error {
	dataMap := map[string]interface{}{
		"birthday": birthday,
	}

	whereMap := map[string]string{
		"card_id": fmt.Sprintf("%d", t.CardId),
	}

	sql := db.Update(tableName, whereMap, dataMap)
	stmt, err := db.MysqlHandle.Prepare(sql)
	if err != nil {
		godog.Error("errors occur while util.Db_zone.Prepare(): %s", err.Error())
		return err
	}

	defer stmt.Close()

	res, err := stmt.Exec(birthday)
	if err != nil {
		godog.Error("errors occur while stmt.Exec(): %s", err.Error())
		return err
	}

	num, err := res.RowsAffected()
	if err != nil {
		godog.Error("errors occur while res.RowsAffected(): %s", err.Error())
		return err
	}

	if num != 1 {
		return errors.New("none row Affected")
	}

	return nil
}

func (t *Test) Query(cardId uint64) (*Test, error) {
	sql := `SELECT name,sex,birthday,status,create_time FROM ` + tableName + ` WHERE card_id = ? `

	rows, err := db.MysqlHandle.Query(sql, cardId)
	if err != nil {
		godog.Error("occur error :%s", err)
		return nil, err
	}

	defer rows.Close()

	var app *Test = nil
	for rows.Next() {
		app = &Test{}
		app.CardId = cardId
		err = rows.Scan(&app.Name, &app.Sex, &app.Birthday, &app.Status, &app.CreateTs)
		if err != nil {
			godog.Error("occur error :%s", err)
			return nil, err
		}
	}

	return app, nil
}

func testAdd() {
	t := &Test{
		Name:     "chuck",
		CardId:   1025,
		Sex:      "male",
		Birthday: 1024,
		Status:   1,
		CreateTs: uint64(time.Now().Unix()),
	}

	if err := t.Add(); err != nil {
		godog.Error("[testAdd] errors occur while res.RowsAffected(): %s", err.Error())
		return
	}
}

func testUpdate() {
	t := &Test{
		CardId: 1024,
	}

	if err := t.Update(1025); err != nil {
		godog.Error("[testUpdate] errors occur while res.RowsAffected(): %s", err.Error())
		return
	}
}

func testQuery() {
	t := &Test{}

	tt, err := t.Query(1024)
	if err != nil {
		godog.Error("query occur error:", err)
		return
	}

	godog.Debug("query: %v", *tt)
}

func main() {
    testAdd()
	testQuery()
	testUpdate()
	testQuery()
}
```

```
/**
 * Copyright 2018 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"godog"
	"godog/store/cache"
)

func main() {
	key := "key"
	if err := cache.RedisHandle.Set(key, "value", 10, 0, false, true); err != nil {
		godog.Error("redis set occur error:%s", err)
		return
	}

	value, err := cache.RedisHandle.Get(key)
	if err != nil {
		godog.Error("redis get occur error:%s", err)
		return
	}

	godog.Debug("value:%s", string(value))
}
```
## License

Godog is released under the [**MIT LICENSE**](http://opensource.org/licenses/mit-license.php).  

