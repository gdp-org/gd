/**
 * Copyright 2020 gl Author. All rights reserved.
 * Author: Chuck1024
 */

package gl

const shardCount = 1024

type goroutineLocal struct {
	m *ConcurrentMap
}

func newGoroutineLocal() *goroutineLocal {
	return &goroutineLocal{m: NewShard(shardCount)}
}

func (g *goroutineLocal) getGl() (map[interface{}]interface{}, bool) {
	goId, ok := getGoId()
	if !ok {
		return nil, false
	}
	glObj, ok := g.m.Get(goId)
	if !ok || glObj == nil {
		return nil, false
	}
	gl, ok := glObj.(map[interface{}]interface{})
	if ok {
		return gl, ok
	} else {
		return nil, false
	}
}
