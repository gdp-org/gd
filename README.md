# godog

"go" is the meaning of a dog in Chinese pronunciation, and dog's original intention is also a dog. So godog means "狗狗" in Chinese, which is very cute.


## Author

```
author:Chuck1024
email:chuck.ch1024@outlook.com
```

## Installation

Start with cloning godog:

```
>git clone https://github.com/chuck1024/godog.git
```

##Introduction

Godog is a basic framework with golang, which can help developers develop server quickly. It provides rich modules to help you.

It contains `config module`,`error module`,`logging module`,`net module` and `service module`. You can use any of the modules according to your needs. More modules will be added later. I hope you can join in and we make it more richer and useful that it can be used by more projects.

>* [logging](https://github.com/xuyu/logging)  module is third-party library. Author is [**xuyu**](https://github.com/xuyu). Thanks for xuyu here.  

## Usage

`service module` provides golang server. It is a simple demo that you can develop it on the basis of it.

```
func HandlerTest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test success!!!"))
}

func main() {
	App := service.NewApplication("test")
	App.AddHandlerFunc("/test", HandlerTest)

	err := App.Run()
	if err != nil {
		logging.Error("Error occurs, error = %s", err.Error())
		return
	}
}

// you can use command to test service that it is in another file <serviceTest.txt>.
```

`config module` provides the related configuration of the project.

```
func TestConfig(t *testing.T){
	AppConfig = config.AppConfig

	// Notice: config contains BaseConfigure. config.json must contain the BaseConfigure configuration.
	// The location of config.json is "conf/conf.json". Of course, you change it if you want.

	// AppConfig.BaseConfig.Log.File is the path of log file.
	file := AppConfig.BaseConfig.Log.File
	logging.Debug("log file:%s",file)

	// AppConfig.BaseConfig.Log.Level is log level.
	// DEBUG   logLevel = 1
	// INFO    logLevel = 2
	// WARNING logLevel = 3
	// ERROR   logLevel = 4
	// DISABLE logLevel = 255
	level := AppConfig.BaseConfig.Log.Level
	logging.Debug("log level:%s",level)

	// AppConfig.BaseConfig.Log.Name is service name
	name := AppConfig.BaseConfig.Log.Name
	logging.Debug("name:%s",name)

	// AppConfig.BaseConfig.Log.Suffix is suffix of log file.
	// suffix = "060102-15" . It indicates that the log is cut per hour
	// suffix = "060102" . It indicates that the log is cut per day
	suffix := AppConfig.BaseConfig.Log.Suffix
	logging.Debug("log suffix:%s",suffix)

	// you can add configuration items directly in conf.json
	value := AppConfig.Get("key")
	logging.Debug("value:%s",value)

	// you can add config key-value if you need.
	AppConfig.Set("yourKey","yourValue")

	// get config key
	yourValue := AppConfig.Get("yourKey")
	logging.Debug("yourValue:%s",yourValue)
}
```

`error module` provides the relation usages of error which you find in getResponseInfo function of “net/http/httplib.go”

## License

Godog is released under the [**MIT LICENSE**](http://opensource.org/licenses/mit-license.php).  

