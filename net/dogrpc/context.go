/**
 * Copyright 2019 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package dogrpc

type Context struct {
	Seq          uint32
	//Service      string
	Method       string
	//Args         map[string]interface{}
	Handler      Handler
	Req          []byte
}