/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/gdp-org/gd"
	"github.com/gdp-org/gd/databases/mysqldb"
)

type TestDB struct {
	Id       uint64 `json:"id" mysqlField:"id" `
	Name     string `json:"name" mysqlField:"name"`
	CardId   uint64 `json:"card_id" mysqlField:"card_id" `
	Sex      string `json:"sex" mysqlField:"sex" dataType:"clob"`
	Birthday uint64 `json:"birthday" mysqlField:"birthday"`
	Status   uint64 `json:"status" mysqlField:"status"`
	CreateTs uint64 `json:"create_time" mysqlField:"create_time"`
	UpdateTs uint64 `json:"update_time" mysqlField:"update_time"`
}

func main() {
	defer gd.LogClose()
	gd.SetConfPath("H:\\GoProgram\\Go\\src\\gd2\\databases\\mysqldb\\sample\\conf\\conf.ini")
	o := mysqldb.MysqlClient{DataBase: "honeypot", DbConf: gd.GetConfFile()}
	if err := o.Start(); err != nil {
		gd.Error("err:%s", err)
		return
	}

	// Query
	query := "select ? from test where  id=? limit 1"
	var err error

	data, err := o.Query((*TestDB)(nil), query, 1)
	if err != nil {
		gd.Error("err:%s", err)
		return
	}

	if data == nil {
		gd.Error("err:%s", err)
		return
	}
	gd.Debug("%v", data.(*TestDB))

	insert := &TestDB{
		Name:     "chucks",
		CardId:   145251,
		Sex:      "male",
		Birthday: 1312412,
		Status:   1,
		CreateTs: 112131231,
		UpdateTs: 112131231,
	}

	err = o.Add("test", insert, false)
	if err != nil {
		gd.Error("%s", err)
	}

	_, err = o.AddEscapeAutoIncrAndRetLastId("test", insert, "id")
	if err != nil {
		gd.Error("%s", err)
	}

	insert2 := &TestDB{
		Name:     "xxxxxxx",
		CardId:   1111444443333,
		Sex:      "male",
		Birthday: 1312412,
		Status:   1,
		CreateTs: 112131231,
		UpdateTs: 112131231,
	}

	_, err = o.AddEscapeAutoIncr("test", insert2, true, "id")
	if err != nil {
		gd.Error("%s", err)
	}

	_, err = o.AddEscapeAutoIncrAndRetLastId("test", insert2, "id")
	if err != nil {
		gd.Error("%s", err)
	}

	updateData := &TestDB{
		Id:       214,
		Name:     "rqtyhjkl;lgkfjdhg",
		CardId:   4125115,
		Sex:      "1111",
		Birthday: 66666666,
		Status:   1,
		CreateTs: 000000000,
		UpdateTs: 141251,
	}

	primaryKey := []string{"id", "sex"}
	updateFiled := []string{"name", "create_time"}

	_, err = o.InsertOrUpdateOnDup("test", updateData, primaryKey, updateFiled, true)
	if err != nil {
		gd.Error("%s", err)
	}

	// 支持
	c, err := o.GetCount("select count(*) from test") //
	if err != nil {
		gd.Error("%s", err)
	}
	gd.Info("count %d", c)

	query1 := "select ? from test where sex = ? " //
	retList, err := o.QueryList((*TestDB)(nil), query1, "male")
	testList := make([]*TestDB, 0)
	for _, ret := range retList {
		product, _ := ret.(*TestDB)
		testList = append(testList, product)
	}
	gd.Debug("%v", testList[0].Name)

	where := make(map[string]interface{}, 0)
	where["sex"] = ""
	where["id"] = 1
	err = o.Update("test", &TestDB{Birthday: 444444444}, where, []string{"birthday"})
	if err != nil {
		gd.Error("%s", err)
	}

	con := make(map[string]interface{})
	con["name"] = []string{"2222"}
	_, err = o.Delete("test", con)
	if err != nil {
		gd.Error("%s", err)
	}
}
