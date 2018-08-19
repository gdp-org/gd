/**
 * Copyright 2018 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"errors"
	"fmt"
	"github.com/chuck1024/godog"
	"github.com/chuck1024/godog/store/db"
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
	testQuery()
}
