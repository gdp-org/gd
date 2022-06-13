/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/chuck1024/gd"
	"github.com/chuck1024/gd/databases/mysqldb"
)

type TestDB struct {
	Id       uint64 `json:"id" mysqlField:"id"`
	Name     string `json:"name" mysqlField:"name"`
	CardId   uint64 `json:"card_id" mysqlField:"card_id"`
	Sex      string `json:"sex" mysqlField:"sex"`
	Birthday uint64 `json:"birthday" mysqlField:"birthday"`
	Status   uint64 `json:"status" mysqlField:"status"`
	CreateTs uint64 `json:"create_time" mysqlField:"create_time"`
	UpdateTs uint64 `json:"update_time" mysqlField:"update_time"`
}

type TestDB1 struct {
	Id         uint64 `json:"id" mysqlField:"id"`
	Name       string `json:"name" mysqlField:"name"`
	CreateTime int    `json:"createTime" mysqlField:"createTime"`
}

func main() {
	defer gd.LogClose()
	gd.SetConfPath("H:\\GoProgram\\Go\\src\\gd\\databases\\mysqldb\\sample\\conf\\conf.ini")
	o := mysqldb.MysqlClient{DataBase: "honeypot", DbConf: gd.GetConfFile()}
	if err := o.Start(); err != nil {
		gd.Error("err:%s", err)
		return
	}

	// Query
	query := "select ? from test1 where id =?"
	var err error

	// Add
	insert := &TestDB{
		Name:     "chucks",
		CardId:   132124,
		Sex:      "male",
		Birthday: 1312412,
		Status:   1,
		CreateTs: 112131231,
		UpdateTs: 112131231,
	}

	// 支持
	data, err := o.Query((*TestDB)(nil), query, 2)
	if err != nil {
		gd.Error("err:%s", err)
		return
	}
	if data == nil {
		gd.Error("err:%s", err)
		return
	}
	gd.Debug("%v", data.(*TestDB))

	// dm 不支持
	err = o.Add("test", insert, false)
	if err != nil {
		gd.Error("%s", err)
	}

	updateData := &TestDB1{
		Id:         2,
		Name:       "asdasda",
		CreateTime: 13152512,
	}
	primaryKey := []string{"id"}
	updateFiled := []string{"name", "createTime"}

	// false 支持
	_, err = o.InsertOrUpdateOnDup("test1", updateData, primaryKey, updateFiled, false)
	if err != nil {
		gd.Error("%s", err)
	}

	// 支持
	c, err := o.GetCount("select count(*) from test")
	if err != nil {
		gd.Error("%s", err)
	}
	gd.Info("count %d", c)

	// 支持
	_, err = o.AddEscapeAutoIncrAndRetLastId("test", insert, "id")
	if err != nil {
		gd.Error("%s", err)
	}

	// 支持
	query = "select ? from test1 where name = ? "
	retList, err := o.QueryList((*TestDB1)(nil), query, "xianglei")
	testList := make([]*TestDB1, 0)
	for _, ret := range retList {
		product, _ := ret.(*TestDB1)
		testList = append(testList, product)
	}
	gd.Debug("%v", testList[0].Name)

	// 支持
	where := make(map[string]interface{}, 0)
	where["id"] = 2
	err = o.Update("test1", &TestDB1{Name: "xxx"}, where, []string{"name"})
	if err != nil {
		gd.Error("%s", err)
	}

	// 支持
	con := make(map[string]interface{})
	con["name"] = "xxx"
	_, err = o.Delete("test", con)
	if err != nil {
		gd.Error("%s", err)
	}
}
