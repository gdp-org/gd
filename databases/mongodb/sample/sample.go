/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"context"
	"encoding/json"
	"github.com/chuck1024/gd/databases/mongodb"
	"github.com/chuck1024/gd/dlog"
	"go.mongodb.org/mongo-driver/bson"
)

type Trainer struct {
	Name string
	Age  int
	City string
}

func main() {
	var i chan struct{}
	o := mongodb.MongoClient{
		DataBase: "test",
	}

	if err := o.Start(); err != nil {
		dlog.Error("err:%s", err)
		return
	}

	// insert
	testA := &Trainer{
		Name: "testB",
		Age:  18,
		City: "ChongQi",
	}

	result, err := o.Insert("test", []interface{}{testA})
	if err != nil {
		dlog.Error("insert occur error:%v", err)
		return
	}
	dlog.Debug("insert result:%v", result)

	// Find One
	filter := bson.D{{"age", 18}}
	var r1 *Trainer

	result1, err := o.FindOne("test", filter)
	if err != nil {
		dlog.Error("FindOne occur error:%v", err)
		return
	}
	if err = result1.Decode(&r1); err != nil {
		dlog.Error("FindOne  Decodeoccur error:%v", err)
		return
	}
	dlog.Debug("FindOne result:%v", r1)

	// Find Many
	var r2 []*Trainer

	result2, err := o.Find("test", filter)
	if err != nil {
		dlog.Error("Find occur error:%v", err)
		return
	}

	for result2.Next(context.TODO()) {
		var tmp *Trainer
		if err = result2.Decode(&tmp); err != nil {
			dlog.Error("Find Decode occur error:%v", err)
			continue
		}

		r2 = append(r2, tmp)
	}
	r2Str, _ := json.Marshal(r2)
	dlog.Debug("Find result:%v", string(r2Str))

	// Update One
	update := bson.D{
		{"$inc", bson.D{
			{"age", 1},
		}},
	}

	r3, err := o.UpdateOne("test", update, filter)
	if err != nil {
		dlog.Error("UpdateOne occur error:%v", err)
		return
	}
	dlog.Debug("UpdateOne result:%v", r3)

	// Update Many
	r3, err = o.UpdateMany("test", update, filter)
	if err != nil {
		dlog.Error("UpdateMany occur error:%v", err)
		return
	}
	dlog.Debug("UpdateMany result:%v", r3)

	// Delete One
	filter = bson.D{{"age", 19}}
	r4, err := o.DeleteOne("test", filter)
	if err != nil {
		dlog.Error("DeleteOne occur error:%v", err)
		return
	}
	dlog.Debug("DeleteOne result:%v", r4)

	// Delete Many
	r4, err = o.DeleteMany("test", filter)
	if err != nil {
		dlog.Error("DeleteMany occur error:%v", err)
		return
	}
	dlog.Debug("DeleteMany result:%v", r4)

	<-i
}
