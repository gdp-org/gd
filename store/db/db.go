/**
 * Copyright 2018 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xuyu/logging"
	"godog/config"
	"regexp"
	"strings"
)

var (
	MysqlHandle *sql.DB
)

func init() {
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
		logging.Error("[init] Failed mysql url=" + url + ",err=" + err.Error())
		panic("failed mysql url=" + url)
		return
	}

	logging.Debug("[init] maxConnections=%d", maxConnections)
	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(10)

	logging.Info("Mysql conn ok: %s", url)
	MysqlHandle = db
}

func Where(whereMap map[string]string) string {
	whereSql := ""
	isStart := true
	for key, value := range whereMap {
		if isStart == true {
			whereSql = fmt.Sprintf(" `%s` = '%s' ", key, value)
			isStart = false
		} else {

			whereSql = fmt.Sprintf("%s And `%s` = '%s' ", whereSql, key, value)
		}
	}

	if whereSql == "" {
		whereSql = " 1!=1 "
	}
	return whereSql
}

func WhereSafety(whereMap map[string]string) (string, []interface{}) {
	values := make([]interface{}, 0)

	whereSql := ""
	isStart := true
	for key, value := range whereMap {
		if isStart != true {
			whereSql = whereSql + "And"
		}
		whereSql = fmt.Sprintf(" %s `%s` = ? ", whereSql, key)
		values = append(values, value)
		isStart = false
	}

	if whereSql == "" {
		whereSql = " 1!=1 "
	}

	return whereSql, values
}

func InsertOne(tableName string, sqlMap map[string]interface{}) string {
	fields := ""
	sqlData := ""
	isStart := true

	for key := range sqlMap {
		if isStart != true {
			fields = fields + ","
			sqlData = sqlData + ","
		}

		fields = fmt.Sprintf("%s  `%s`", fields, key)
		sqlData = fmt.Sprintf("%s ?", sqlData)
		isStart = false
	}
	sqlString := fmt.Sprintf("INSERT INTO %s (%s) values (%s)", tableName, fields, sqlData)
	logging.Debug("InsertOne sql:%s", sqlString)

	return sqlString
}

func Update(tableName string, whereMap map[string]string, setMap map[string]interface{}) string {
	setSql := ""
	isStart := true

	for key := range setMap {
		if isStart != true {
			setSql = setSql + ","
		}
		setSql = fmt.Sprintf("%s  `%s` = ? ", setSql, key)
		isStart = false
	}

	whereSql := Where(whereMap)

	sqlString := fmt.Sprintf("update %s set %s where %s", tableName, setSql, whereSql)
	logging.Debug("Update sql:%s", sqlString)

	return sqlString
}
