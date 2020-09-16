/**
 * Copyright 2019 redisdb Author. All rights reserved.
 * Author: Chuck1024
 */

package redisdb

import (
	"encoding/json"
	"strconv"
)

func stringInSlice(a []string, s string) bool {
	if a == nil || len(a) == 0 {
		return false
	}
	for _, v := range a {
		if s == v {
			return true
		}
	}
	return false
}

func MustString(v interface{}, defaultValue string) string {
	switch tv := v.(type) {
	case string:
		return tv
	case []byte:
		return string(tv)
	case int64:
		return strconv.FormatInt(int64(tv), 10)
	case uint64:
		return strconv.FormatUint(uint64(tv), 10)
	case int32:
		return strconv.FormatInt(int64(tv), 10)
	case uint32:
		return strconv.FormatUint(uint64(tv), 10)
	case int:
		return strconv.Itoa(int(tv))
	case int16:
		return strconv.FormatInt(int64(tv), 10)
	case uint16:
		return strconv.FormatUint(uint64(tv), 10)
	case float32:
		return strconv.FormatFloat(float64(tv), 'f', -1, 64)
	case float64:
		return strconv.FormatFloat(tv, 'f', -1, 64)
	case json.Number:
		return tv.String()
	case bool:
		if tv {
			return "true"
		} else {
			return "false"
		}
	}
	return defaultValue
}
