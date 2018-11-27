/**
 * Copyright 2018 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package main_test

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/chuck1024/godog"
	"github.com/chuck1024/godog/dao/db"
	"testing"
	"time"
)

const (
	tableName = "test"
)

var (
	MysqlHandle *sql.DB
)

type TestDB struct {
	Name     string `json:"name"`
	CardId   uint64 `json:"card_id"`
	Sex      string `json:"sex"`
	Birthday uint64 `json:"birthday"`
	Status   uint8  `json:"status"`
	CreateTs uint64 `json:"create_time"`
}

func (t *TestDB) Add() error {
	insertData := map[string]interface{}{
		"name":        t.Name,
		"card_id":     t.CardId,
		"sex":         t.Sex,
		"birthday":    t.Birthday,
		"status":      t.Status,
		"create_time": t.CreateTs,
	}

	sqlData := db.InsertOne(tableName, insertData)
	stmt, err := MysqlHandle.Prepare(sqlData)
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

func (t *TestDB) Update(birthday uint64) error {
	dataMap := map[string]interface{}{
		"birthday": birthday,
	}

	whereMap := map[string]string{
		"card_id": fmt.Sprintf("%d", t.CardId),
	}

	sqlData := db.Update(tableName, whereMap, dataMap)
	stmt, err := MysqlHandle.Prepare(sqlData)
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

func (t *TestDB) Query(cardId uint64) (*TestDB, error) {
	sqlData := `SELECT name,sex,birthday,status,create_time FROM ` + tableName + ` WHERE card_id = ? `

	rows, err := MysqlHandle.Query(sqlData, cardId)
	if err != nil {
		godog.Error("occur error :%s", err)
		return nil, err
	}

	defer rows.Close()

	var app *TestDB = nil
	for rows.Next() {
		app = &TestDB{}
		app.CardId = cardId
		err = rows.Scan(&app.Name, &app.Sex, &app.Birthday, &app.Status, &app.CreateTs)
		if err != nil {
			godog.Error("occur error :%s", err)
			return nil, err
		}
	}

	return app, nil
}

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
		t.Logf("[init] get config mysql url occur error: %s", err)
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
		t.Logf("[init] get config mysql url occur error:%s ", err)
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
