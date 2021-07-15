/**
 * Copyright 2018 Author. All rights reserved.
 * Author: Chuck1024
 */

package sample_test

import (
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/service/register"
	"testing"
	"time"
)

func TestEtcd(t *testing.T) {
	var r register.DogRegister
	var i chan struct{}

	r = &register.EtcdRegister{}
	if err := r.Start(); err != nil {
		dlog.Error("err:%s", err)
		return
	}

	time.Sleep(3 * time.Second)
	r.Close()
	<-i
}

func TestZk(t *testing.T) {
	var r register.DogRegister
	var i chan struct{}

	r = &register.ZkRegister{}
	if err := r.Start(); err != nil {
		dlog.Error("err:%s", err)
		return
	}
	time.Sleep(10 * time.Second)
	r.Close()
	<-i
}
