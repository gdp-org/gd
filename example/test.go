/**
 * Created by JetBrains GoLand.
 * Author: Chuck Chen
 * Date: 2018/6/22
 * Time: 16:19
 */

package main

import (
	"github.com/xuyu/logging"
	"godog/config"
	_ "godog/logs"
)

func main() {
	appConfig := config.AppConfig

	logging.Debug(appConfig.Get("File"))

}
