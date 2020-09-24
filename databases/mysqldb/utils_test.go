/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package mysqldb

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMysqlClient_getFieldNames(t *testing.T) {
	Convey("test get field name", t, func() {
		res, err := GetFieldsName((*userSceneAuthed)(nil))
		So(err, ShouldBeNil)
		So(res, ShouldEqual, "`us_id`,`did`,`type`,`create_time`")

		res, err = GetFieldsName(userSceneAuthed{})
		So(err, ShouldBeNil)
		So(res, ShouldEqual, "`us_id`,`did`,`type`,`create_time`")

		res, err = GetFieldsName(nil)
		So(err, ShouldNotBeNil)

		res, err = GetFieldsName("aaa")
		So(err, ShouldNotBeNil)

		res, err = GetFieldsName(&struct{ Name int }{1})
		So(err, ShouldBeNil)
		So(res, ShouldEqual, "")
	})
}

func TestMysqlClient_getFieldNames_pb(t *testing.T) {
	Convey("test get field name", t, func() {
		res, err := GetFieldsName((*device)(nil))
		So(err, ShouldBeNil)
		So(res, ShouldEqual, "`id`,`pd_id`,`token`")

		res, err = GetFieldsName(device{})
		So(err, ShouldBeNil)
		So(res, ShouldEqual, "`id`,`pd_id`,`token`")

		res, err = GetFieldsName(nil)
		So(err, ShouldNotBeNil)

		res, err = GetFieldsName("aaa")
		So(err, ShouldNotBeNil)

		res, err = GetFieldsName(&struct{ Name int }{1})
		So(err, ShouldBeNil)
		So(res, ShouldEqual, "")
	})
}

func TestMysqlClient_getFields_pb(t *testing.T) {
	Convey("test get field", t, func() {
		res, err := GetFields((*device)(nil))
		So(err.Error(), ShouldEqual, "input cannot be nil")
		So(res, ShouldBeNil)

		data := &device{
			Id:    "1231sqq",
			PdId:  984,
			Token: "xx123",
		}
		res, err = GetFields(data)
		So(err, ShouldBeNil)
		So(len(res), ShouldEqual, 3)
		So(res[0], ShouldEqual, &data.Id)
		So(res[1], ShouldEqual, &data.PdId)
		So(res[2], ShouldEqual, &data.Token)
	})
}

func TestMysqlClient_getFields(t *testing.T) {
	Convey("test get field", t, func() {
		res, err := GetFields((*userSceneAuthed)(nil))
		So(err.Error(), ShouldEqual, "input cannot be nil")
		So(res, ShouldBeNil)

		data := &userSceneAuthed{
			UsId:       123,
			Did:        "123",
			Type:       456,
			CreateTime: 789,
		}
		res, err = GetFields(data)
		So(err, ShouldBeNil)
		So(len(res), ShouldEqual, 4)
		So(res[0], ShouldEqual, &data.UsId)
		So(res[1], ShouldEqual, &data.Did)
		So(res[2], ShouldEqual, &data.Type)
		So(res[3], ShouldEqual, &data.CreateTime)
	})
}

func BenchmarkMysqlClient_getFieldsName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		res, err := GetFieldsName((*userSceneAuthed)(nil))
		if err != nil {
			b.Fatalf("err should be nil")
		}
		if res != "us_id,did,type,create_time" {
			b.Fatalf("result failed")
		}
	}
}
func BenchmarkMysqlClient_getFields(b *testing.B) {
	for i := 0; i < b.N; i++ {
		data := &userSceneAuthed{
			UsId:       123,
			Did:        "123",
			Type:       456,
			CreateTime: 789,
		}
		_, err := GetFields(data)
		if err != nil {
			fmt.Errorf("err not nil")
		}
	}
}

type device struct {
	Id    string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	PdId  int32  `protobuf:"varint,2,opt,name=pd_id,json=pdId" json:"pd_id,omitempty"`
	Token string `protobuf:"bytes,3,opt,name=token" json:"token,omitempty"`
}

type userSceneAuthed struct {
	UsId       int64  `mysqlField:"us_id"`
	Did        string `mysqlField:"did"`
	Type       int64  `mysqlField:"type"`
	CreateTime int64  `mysqlField:"create_time"`
}

func TestMysqlClient_buildWhereSql(t *testing.T) {
	Convey("test build where str", t, func() {
		condition := make(map[string]interface{})
		condition["uid"] = 1
		condition["name"] = "123"
		condition["int64s"] = []int64{1, 3, 4, 6, 7}
		condition["strings"] = []string{"1", "3", "4", "6", "7"}
		condition["interfaces"] = []interface{}{1, 3, 4, 6, 7}
		str, val := buildWhereSql(condition)
		parts := strings.Split(str, "AND")
		curridx := 0
		So(len(parts), ShouldEqual, 5)
		for k := 0; k < len(parts); k++ {
			v := parts[k]
			if strings.Contains(v, "uid") {
				So(strings.Trim(v, " "), ShouldEqual, "`uid` = ?")
				So(val[curridx], ShouldEqual, 1)
				curridx++
				continue
			}
			if strings.Contains(v, "name") {
				So(strings.Trim(v, " "), ShouldEqual, "`name` = ?")
				So(val[curridx], ShouldEqual, "123")
				curridx++
				continue
			}
			if strings.Contains(v, "int64s") {
				So(strings.Trim(v, " "), ShouldEqual, "`int64s` in (?,?,?,?,?)")
				So(val[curridx], ShouldEqual, 1)
				So(val[curridx+1], ShouldEqual, 3)
				So(val[curridx+2], ShouldEqual, 4)
				So(val[curridx+3], ShouldEqual, 6)
				So(val[curridx+4], ShouldEqual, 7)
				curridx = curridx + 5
				continue
			}
			if strings.Contains(v, "interfaces") {
				So(strings.Trim(v, " "), ShouldEqual, "`interfaces` in (?,?,?,?,?)")
				So(val[curridx], ShouldEqual, 1)
				So(val[curridx+1], ShouldEqual, 3)
				So(val[curridx+2], ShouldEqual, 4)
				So(val[curridx+3], ShouldEqual, 6)
				So(val[curridx+4], ShouldEqual, 7)
				curridx = curridx + 5
				continue
			}
			if strings.Contains(v, "strings") {
				So(strings.Trim(v, " "), ShouldEqual, "`strings` in (?,?,?,?,?)")
				So(val[curridx], ShouldEqual, "1")
				So(val[curridx+1], ShouldEqual, "3")
				So(val[curridx+2], ShouldEqual, "4")
				So(val[curridx+3], ShouldEqual, "6")
				So(val[curridx+4], ShouldEqual, "7")
				curridx = curridx + 5
				continue
			}
		}
	})

}

func TestCondtion(t *testing.T) {
	Convey("test condition ", t, func() {
		Convey("test new condition build", func() {
			c := NewSqlCondition()
			tablename, str, vars := c.BuildShardWhereSql("")
			So(tablename, ShouldEqual, "")
			So(strings.Trim(str, " "), ShouldEqual, "LIMIT ?")
			So(vars[0], ShouldEqual, 300)
		})
		Convey("test with tableprefix", func() {
			c := NewSqlCondition()
			c.WithTablePrefix("test")
			tablename, str, vars := c.BuildShardWhereSql("123")
			So(tablename, ShouldEqual, "test123")
			So(strings.Trim(str, " "), ShouldEqual, "LIMIT ?")
			So(vars[0], ShouldEqual, 300)
		})
		Convey("test with condition", func() {
			c := NewSqlCondition()
			c.WithCondition("test1", ">", 123)
			tablename, str, vars := c.BuildShardWhereSql("")
			So(tablename, ShouldEqual, "")
			So(strings.Trim(str, " "), ShouldEqual, "WHERE `test1` > ? LIMIT ?")
			So(vars[0], ShouldEqual, 123)
			So(vars[1], ShouldEqual, 300)
		})
		Convey("test mix condition", func() {
			c := NewSqlCondition()
			c.WithTablePrefix("testtable")
			c.WithCondition("test1", ">", 1)
			c.WithCondition("test2", "=", 2)
			c.WithCondition("test3", "", "test")
			c.WithCondition("test4", "", []int64{4, 5, 6, 7})
			c.WithOrder("test1", true)
			c.WithOrder("test2", false)
			c.WithLimit(100)
			c.WithOffset(200)
			tablename, str, vars := c.BuildShardWhereSql("123")
			So(tablename, ShouldEqual, "testtable123")
			So(strings.Trim(str, " "), ShouldEqual, "WHERE `test1` > ? AND `test2` = ? AND `test3` = ? AND `test4` IN (?,?,?,?) ORDER BY `test1` DESC,`test2` ASC LIMIT ?,?")
			So(len(vars), ShouldEqual, 9)
			So(vars[0], ShouldEqual, 1)
			So(vars[1], ShouldEqual, 2)
			So(vars[2], ShouldEqual, "test")
			So(vars[3], ShouldEqual, 4)
			So(vars[4], ShouldEqual, 5)
			So(vars[5], ShouldEqual, 6)
			So(vars[6], ShouldEqual, 7)
			So(vars[7], ShouldEqual, 200)
			So(vars[8], ShouldEqual, 100)
		})
		Convey("test mix condition without offset", func() {
			c := NewSqlCondition()
			c.WithTablePrefix("testtable")
			c.WithCondition("test1", ">", 1)
			c.WithCondition("test2", "=", 2)
			c.WithCondition("test3", "", "test")
			c.WithCondition("test4", "", []int64{4, 5, 6, 7})
			c.WithOrder("test1", true)
			c.WithOrder("test2", false)
			c.WithLimit(100)
			tablename, str, vars := c.BuildShardWhereSql("123")
			So(tablename, ShouldEqual, "testtable123")
			So(strings.Trim(str, " "), ShouldEqual, "WHERE `test1` > ? AND `test2` = ? AND `test3` = ? AND `test4` IN (?,?,?,?) ORDER BY `test1` DESC,`test2` ASC LIMIT ?")
			So(len(vars), ShouldEqual, 8)
			So(vars[0], ShouldEqual, 1)
			So(vars[1], ShouldEqual, 2)
			So(vars[2], ShouldEqual, "test")
			So(vars[3], ShouldEqual, 4)
			So(vars[4], ShouldEqual, 5)
			So(vars[5], ShouldEqual, 6)
			So(vars[6], ShouldEqual, 7)
			So(vars[7], ShouldEqual, 100)
		})
		Convey("test mix condition without limit", func() {
			c := NewSqlCondition()
			c.WithTablePrefix("testtable")
			c.WithCondition("test1", ">", 1)
			c.WithCondition("test2", "=", 2)
			c.WithCondition("test3", "", "test")
			c.WithCondition("test4", "", []int64{4, 5, 6, 7})
			c.WithOrder("test1", true)
			c.WithOrder("test2", false)
			c.WithLimit(0)
			tablename, str, vars := c.BuildShardWhereSql("123")
			So(tablename, ShouldEqual, "testtable123")
			So(strings.Trim(str, " "), ShouldEqual, "WHERE `test1` > ? AND `test2` = ? AND `test3` = ? AND `test4` IN (?,?,?,?) ORDER BY `test1` DESC,`test2` ASC")
			So(len(vars), ShouldEqual, 7)
			So(vars[0], ShouldEqual, 1)
			So(vars[1], ShouldEqual, 2)
			So(vars[2], ShouldEqual, "test")
			So(vars[3], ShouldEqual, 4)
			So(vars[4], ShouldEqual, 5)
			So(vars[5], ShouldEqual, 6)
			So(vars[6], ShouldEqual, 7)
		})
		Convey("test build with empty", func() {
			c := NewSqlCondition()
			c.WithTablePrefix("testtable")
			c.WithLimit(0)
			tablename, str, vars := c.BuildShardWhereSql("123")
			So(tablename, ShouldEqual, "testtable123")
			So(str, ShouldEqual, "")
			So(len(vars), ShouldEqual, 0)
		})
	})
}

