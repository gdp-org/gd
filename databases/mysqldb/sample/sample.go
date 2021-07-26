/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"github.com/chuck1024/gd/databases/mysqldb"
	"github.com/chuck1024/gd/dlog"
)

type TestDB struct {
	Id       uint64  `json:"id" mysqlField:"id"`
	Name     string  `json:"name" mysqlField:"name"`
	CardId   uint64  `json:"card_id" mysqlField:"card_id"`
	Sex      string  `json:"sex" mysqlField:"sex"`
	Birthday uint64  `json:"birthday" mysqlField:"birthday"`
	Status   uint8   `json:"status" mysqlField:"status"`
	CreateTs uint64  `json:"create_time" mysqlField:"create_time"`
	UpdateTs []uint8 `json:"update_time" mysqlField:"update_time"`
}

func main() {
	defer dlog.Close()
	o := mysqldb.MysqlClient{DataBase: "test"}
	if err := o.Start(); err != nil {
		dlog.Error("err:%s", err)
		return
	}

	// Query
	query := "select ? from test where id = ?"
	data, err := o.Query((*TestDB)(nil), query, 2)
	if err != nil {
		dlog.Error("err:%s", err)
		return
	}
	if data == nil {
		dlog.Error("err:%s", err)
		return
	}
	dlog.Debug("%v", data.(*TestDB))

	// Add
	insert := &TestDB{
		Name:     "chucks",
		CardId:   1026,
		Sex:      "male",
		Birthday: 19991010,
		Status:   1,
	}

	err = o.Add("test", insert, true)
	if err != nil {
		dlog.Error("%s", err)
	}

	// queryList
	query = "select ? from test where name = ? "
	retList, err := o.QueryList((*TestDB)(nil), query, "chucks")
	testList := make([]*TestDB, 0)
	for _, ret := range retList {
		product, _ := ret.(*TestDB)
		testList = append(testList, product)
	}
	dlog.Debug("%v", testList[0].CardId)

	// update
	where := make(map[string]interface{}, 0)
	where["id"] = 2
	err = o.Update("test", &TestDB{Sex: "female"}, where, []string{"sex"})
	if err != nil {
		dlog.Error("%s", err)
	}
}
