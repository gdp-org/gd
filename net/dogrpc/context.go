/**
 * Copyright 2019 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package dogrpc

type Context struct {
	ClientAddr string
	Seq        uint32
	Method     string
	Handler    RpcHandlerFunc
	Req        []byte
}
