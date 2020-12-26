/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package gl

import (
	"fmt"
	js "github.com/bitly/go-simplejson"
	"os"
	"reflect"
	"strings"
	"time"
)

var _gl = newGoroutineLocal()

func Init() {
	goId, ok := getGoId()
	if !ok {
		return
	}
	glObj, ok := _gl.m.Get(goId)
	if !ok || glObj == nil {
		glObj = make(map[interface{}]interface{})
		_gl.m.Set(goId, glObj)
		if log != nil {
			log.Debug("init gl goId: %s", goId)
		}
	} else {
		glObj = make(map[interface{}]interface{})
		_gl.m.Set(goId, glObj)
		if log != nil {
			log.Error("double INIT!init replace gl for goId: %s", goId)
		} else {
			msg := fmt.Sprintf("double INIT!init replace gl for goId: %s", goId)
			fmt.Fprint(os.Stderr, msg)
		}
	}
}

func Close() {
	goId, ok := getGoId()
	if !ok {
		return
	}
	glObj, ok := _gl.getGl()
	if !ok {
		return
	}
	if ok {
		if glObj != nil {
			for k := range glObj {
				delete(glObj, k)
			}
		}
		_gl.m.Remove(goId)
		if log != nil {
			log.Debug("clear gl goId:%s", goId)
		}
	}
}

//Exist find out if current ctx has been init
func Exist() bool {
	_, ok := _gl.getGl()
	return ok
}

func Del(key interface{}) {
	cc, ok := _gl.getGl()
	if !ok {
		return
	}
	delete(cc, key)
}

func Get(key interface{}) (interface{}, bool) {
	cc, ok := _gl.getGl()
	if !ok {
		return nil, false
	}
	ret, ok := cc[key]
	return ret, ok
}

func BatchGet(keys []interface{}) map[interface{}]interface{} {
	cc, ok := _gl.getGl()
	if !ok {
		return nil
	}
	ret := make(map[interface{}]interface{})
	for _, k := range keys {
		v, ok := cc[k]
		if ok {
			ret[k] = v
		}
	}
	return ret
}

func Set(key interface{}, val interface{}) {
	cc, ok := _gl.getGl()
	if !ok {
		return
	}

	cc[key] = val
}

func GetCurrentGlData() map[string]interface{} {
	ret := make(map[string]interface{})
	gid, ok := getGoId()
	if !ok {
		ret["info"] = "no id"
		return ret
	}

	gl, ok := _gl.getGl()
	if !ok || gl == nil {
		ret["info"] = "no gl"
		return ret
	}

	if log != nil {
		log.Debug("gid,gl = %s:%v", gid, gl)
	}

	for k, v := range gl {
		kStr := fmt.Sprintf("%v", k)
		if kStr == ClientIp || kStr == Tag || kStr == LogId || kStr == Url {
			continue
		}
		ret[kStr] = v
	}
	return ret
}

func JsonCurrentCtx() *js.Json {
	retJ := js.New()
	gid, ok := getGoId()
	if !ok {
		retJ.Set("_info", "no id")
		return retJ
	}
	ctx, ok := _gl.getGl()
	if !ok || ctx == nil {
		retJ.Set("_info", "no ctx")
		return retJ
	}

	if log != nil {
		log.Debug("json ctx %s:%v", gid, ctx)
	}

	for k, v := range ctx {
		kStr := fmt.Sprintf("%v", k)
		if kStr == ClientIp || kStr == Tag || kStr == LogId || kStr == Url {
			continue
		}
		if isPtrOrInterface(v) {
			if strings.HasPrefix(kStr, "_tk_") {
				retJ.Set("_tk_", "@#")
			} else {
				retJ.Set(kStr, "@#")
			}
		} else {
			retJ.Set(kStr, v)
		}
	}
	return retJ
}

func isPtrOrInterface(v interface{}) bool {
	if v == nil {
		return true
	}
	t := reflect.TypeOf(v)
	k := t.Kind()
	return k == reflect.Ptr || k == reflect.Interface
}

func IncrCost(key interface{}, cost time.Duration) int64 {
	return Incr(key, int64(cost/time.Millisecond))
}

func IncrCostKey(key string, cost time.Duration) int64 {
	return Incr(key+"_cost", int64(cost/time.Millisecond))
}

func IncrCountKey(key string, val int64) int64 {
	return Incr(key+"_count", val)
}

func IncrFailKey(key string, val int64) int64 {
	return Incr(key+"_fail", val)
}

func Incr(key interface{}, count int64) int64 {
	cc, ok := _gl.getGl()
	if !ok {
		return -1
	}

	v, ok := cc[key]
	if !ok {
		cc[key] = count
		return 0
	}
	vc, ok := v.(int64)
	if !ok {
		cc[key] = count
		return 0
	}

	cc[key] = vc + count
	return vc
}

func Decr(key interface{}, count int64) int64 {
	cc, ok := _gl.getGl()
	if !ok {
		return -1
	}

	v, ok := cc[key]
	if !ok {
		cc[key] = 0 - count
		return 0
	}
	vc, ok := v.(int64)
	if !ok {
		cc[key] = 0 - count
		return 0
	}

	cc[key] = vc - count
	return vc
}
