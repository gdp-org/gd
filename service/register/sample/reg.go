package main

import (
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/service/register"
	"time"
)

func etcd(){
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

func zk(){
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

func main(){
	etcd()
}
