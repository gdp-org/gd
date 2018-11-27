/**
 * Copyright 2018 Author. All rights reserved.
 * Author: Chuck1024
 */

package register

import (
	"testing"
	"time"
)

func TestEtcd(t *testing.T){
	var r DogRegister
	r = &EtcdRegister{}
	r.NewRegister([]string{"localhost:2379"}, "/root/", "stagging","github", "godog", )

	r.Run("127.0.0.1", 10240,10)
	time.Sleep(3 * time.Second)
	r.Close()
}

func TestZk(t *testing.T){
	var r DogRegister
	r = &ZkRegister{}
	r.NewRegister([]string{"localhost:2181"}, "/root/", "stagging","github", "godog", )
	r.Run("127.0.0.1", 10240,10)
	time.Sleep(10 * time.Second)
	r.Close()
}