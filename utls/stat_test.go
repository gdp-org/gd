/**
 * Copyright 2019 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package utls

import (
	"errors"
	"fmt"
	dogErr "github.com/chuck1024/godog/error"
	"log"
	"math/rand"
	"testing"
	"time"
)

func Test_Stat(t *testing.T) {
	StatMgrInstance().Init("stat.log", time.Second*5)
	go func() {
		for i := 0; i < 100; i++ {
			st := NewStat()
			st.Begin("a.b." + string('a'+rand.Intn(10)))

			time.Sleep(time.Millisecond * time.Duration(rand.Intn(150)))

			ret := rand.Intn(3)
			st.End(ret)
		}
	}()

	go func() {
		for i := 0; i < 100; i++ {
			st := NewStat()
			st.Begin("a.b." + string('a'+rand.Intn(10)))

			time.Sleep(time.Millisecond * time.Duration(rand.Intn(1500)))

			st.End(0)
		}
	}()

	time.Sleep(time.Second * time.Duration(61))
	fmt.Println("done.")
}

func test_stat() {
	cmd := "cmd.test"
	begin := GetCurrentMicrosecond()
	ret := 0

	defer func(begin int64) {
		st := NewStat()
		st.BeginAt(cmd, time.Now())
		st.End(ret)
	}(begin)

	ret = 100
	time.Sleep(time.Second)
}

func test_simple_app_stat() *dogErr.CodeError {
	var err *dogErr.CodeError
	defer DeferAppStat(20005, time.Now(), &err)

	//err = NewCodeError(222222, "1111111111")
	time.Sleep(time.Millisecond * 200)

	return err
}

func test_simple_stat() *dogErr.CodeError {
	var err *dogErr.CodeError
	defer DeferStat(time.Now(), &err)

	//err = NewCodeError(222222, "1111111111")
	time.Sleep(time.Millisecond * 200)

	return err
}

func Test_Stat2(t *testing.T) {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	StatMgrInstance().Init("stat.log", time.Second)

	test_simple_stat()

	test_simple_app_stat()

	var err0 error
	var err1 *dogErr.CodeError
	err2 := errors.New("xxx")
	err3 := dogErr.NewCodeError(3, "3")
	var err4 error = nil

	NewStat().Begin("nil").EndErr(nil)
	NewStat().Begin("err0").EndErr(err0)
	NewStat().Begin("err1").EndErr(err1)
	NewStat().Begin("err2").EndErr(err2)
	NewStat().Begin("err3").EndErr(err3)
	NewStat().Begin("err4").EndErr(err4)

	fmt.Printf("%v\n", err0)

	time.Sleep(time.Second * 2)

	return
}
