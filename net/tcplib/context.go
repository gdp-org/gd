/**
 * Copyright 2019 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package tcplib

type Context struct {
	//Service      string
	//Method       string
	//Args         map[string]interface{}
	Handler      Handler
	Req          []byte
}