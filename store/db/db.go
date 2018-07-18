/**
 * Copyright 2018 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package db

import (
	"database/sql"
	"regexp"
	"strings"
	"godog/config"
	"github.com/xuyu/logging"
	_ "github.com/go-sql-driver/mysql"
)

var (
	MysqlHandle *sql.DB
)

func init(){
	url, err := config.AppConfig.String("mysql")
	if err != nil {
		logging.Warning("[init] get config mysql url occur error: ", err)
		return
	}

	maxConnections, err := config.AppConfig.Int("mysqlMaxConn")
	if err != nil {
		logging.Warning("[init] get config mysqlMaxConn occur error: ", err)
		return
	}

	if ok, err := regexp.MatchString("^mysql://.*:.*@.*/.*$", url); !ok || err != nil {
		logging.Error("[init] Mysql config syntax err:mysql_zone,%s,shutdown", url)
		panic("conf error")
		return
	}

	url = strings.Replace(url, "mysql://", "", 1)
	db, err := sql.Open("mysql", url)
	if err != nil {
		logging.Error("[init]Failed mysql url=" + url + ",err=" + err.Error())
		panic("failed mysql url=" + url)
		return
	}

	logging.Debug("[init] maxConnections=%d", maxConnections)
	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(10)

	logging.Info("Mysql conn ok: %s", url)
	MysqlHandle = db
}