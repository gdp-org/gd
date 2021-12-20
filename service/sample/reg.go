package main

import (
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/service/register"
)

func etcd(){
	var r register.DogRegister
	r = &register.EtcdRegister{}
	if err := r.Start(); err != nil {
		dlog.Error("err:%s", err)
		return
	}
}

func zk(){
	var r register.DogRegister
	r = &register.ZkRegister{}
	if err := r.Start(); err != nil {
		dlog.Error("err:%s", err)
		return
	}
}

func main(){
	defer dlog.Close()
	etcd()
}
