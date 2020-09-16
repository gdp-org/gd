/**
 * Copyright 2020 inject Author. All rights reserved.
 * Author: Chuck1024
 */

package inject

import (
	"fmt"
	"testing"
)

type Testing struct {
	Target int `inject:"target"`
}

func (t *Testing) Start() error {
	fmt.Println("start", t.Target)
	return nil
}

func (t *Testing) Close() {
	fmt.Println("close", t.Target)
}

type Dependency struct {
	Test *Testing `inject:"test"`
}

func (d *Dependency) Close() {
	fmt.Println("close Dep", d.Test)
}

func TestExample(t *testing.T) {
	InitDefault()
	//dep.Close, test.Close will be called orderly
	defer Close()
	Reg("target", 123)
	//test will be auto created, test.Start will be called, then dep.Start(if any)
	dep := Reg("dep", (*Dependency)(nil)).(*Dependency)
	fmt.Println("find dep", dep)
}
