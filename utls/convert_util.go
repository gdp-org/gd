package utls

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"

	"fmt"

	"github.com/go-errors/errors"
)

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

func TryString(v interface{}) (string, bool) {
	switch tv := v.(type) {
	case string:
		return tv, true
	case []byte:
		return string(tv), true
	case int64:
		return strconv.FormatInt(int64(tv), 10), true
	case uint64:
		return strconv.FormatUint(uint64(tv), 10), true
	case int32:
		return strconv.FormatInt(int64(tv), 10), true
	case uint32:
		return strconv.FormatUint(uint64(tv), 10), true
	case float32:
		return strconv.FormatFloat(float64(tv), 'f', -1, 64), true
	case float64:
		return strconv.FormatFloat(float64(tv), 'f', -1, 64), true
	case int:
		return strconv.Itoa(int(tv)), true
	case json.Number:
		return tv.String(), true
	case bool:
		if tv {
			return "true", true
		} else {
			return "false", true
		}
	}
	return "", false
}

func MustInt64(v interface{}, defaultValue int64) int64 {
	if v == nil {
		return defaultValue
	}
	switch tv := v.(type) {
	case []byte:
		res, err := strconv.ParseInt(string(tv), 10, 0)
		if err != nil {
			return defaultValue
		}
		return res
	case string:
		res, err := strconv.ParseInt(tv, 10, 0)
		if err != nil {
			return defaultValue
		}
		return res
	case int64:
		return tv
	case uint64:
		if tv > uint64(math.MaxInt64) {
			return defaultValue
		}
		return int64(tv)
	case int32:
		return int64(tv)
	case uint32:
		return int64(tv)
	case int:
		return int64(tv)
	case float32:
		if tv > float32(math.MaxInt64) {
			return defaultValue
		}
		return int64(tv)
	case float64:
		if tv > float64(math.MaxInt64) {
			return defaultValue
		}
		return int64(tv)
	case json.Number:
		val, err := tv.Int64()
		if err == nil {
			return val
		}
	}
	return defaultValue
}

func MustFloat64(v interface{}, defaultValue float64) float64 {
	switch tv := v.(type) {
	case []byte:
		res, err := strconv.ParseFloat(string(tv), 0)
		if err != nil {
			return defaultValue
		}
		return res
	case string:
		res, err := strconv.ParseFloat(tv, 0)
		if err != nil {
			return defaultValue
		}
		return res
	case int64:
		return float64(tv)
	case uint64:
		return float64(tv)
	case int32:
		return float64(tv)
	case uint32:
		return float64(tv)
	case int:
		return float64(tv)
	case float32:
		return float64(tv)
	case float64:
		return tv
	case json.Number:
		val, err := tv.Float64()
		if err == nil {
			return val
		}
	}
	return defaultValue
}

// This is a safe convert since string can always convert to interface{}
func StringStringMap2StringInterfaceMap(input map[string]string) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range input {
		res[k] = v
	}
	return res
}

func ConvertToInt64(v interface{}) (int64, error) {
	defaultValue := int64(-1)
	if v == nil {
		return -1, errors.New("input is nil")
	}
	switch tv := v.(type) {
	case []byte:
		res, err := strconv.ParseInt(string(tv), 10, 0)
		if err != nil {
			return defaultValue, err
		}
		return res, nil
	case string:
		res, err := strconv.ParseInt(tv, 10, 0)
		if err != nil {
			return defaultValue, err
		}
		return res, nil
	case int64:
		return tv, nil
	case uint64:
		if tv > uint64(math.MaxInt64) {
			return defaultValue, errors.New("input number out of range")
		}
		return int64(tv), nil
	case int32:
		return int64(tv), nil
	case uint32:
		return int64(tv), nil
	case int:
		return int64(tv), nil
	case float32:
		if tv > float32(math.MaxInt64) {
			return defaultValue, errors.New("input number out of range")
		}
		return int64(tv), nil
	case float64:
		if tv > float64(math.MaxInt64) {
			return defaultValue, errors.New("input number out of range")
		}
		return int64(tv), nil
	case json.Number:
		val, err := tv.Int64()
		if err == nil {
			return val, nil
		}
		return defaultValue, err
	}
	return defaultValue, errors.Errorf("input number type err type=%v", reflect.TypeOf(v))
}

func SliceCutter(v interface{}, maxsize int) ([]interface{}, error) {
	var ret []interface{}
	x := reflect.ValueOf(v)
	switch x.Kind() {
	case reflect.Array, reflect.String, reflect.Slice:
		length := x.Len()
		turns := length / maxsize
		for i := 0; i <= turns; i++ {
			var part reflect.Value
			if i == turns {
				if length%maxsize != 0 {
					part = x.Slice(maxsize*i, maxsize*i+length%maxsize)
				} else {
					continue
				}
			} else {
				part = x.Slice(maxsize*i, maxsize*(i+1))
			}
			ret = append(ret, part.Interface())
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("not slice types,kind is %v", x.Kind())
	}

}

func MustStringArray(v interface{}, defaultValue []string) []string {
	var ret []string
	x := reflect.ValueOf(v)
	switch x.Kind() {
	case reflect.Array, reflect.String, reflect.Slice:
		for i := 0; i < x.Len(); i++ {
			val := x.Index(i).Interface()
			if val == nil {
				return defaultValue
			}
			if valStr, ok := TryString(val); ok {
				ret = append(ret, valStr)
			} else {
				return defaultValue
			}
		}
		return ret
	default:
		return defaultValue
	}
}

func MustInt64Array(v interface{}, defaultValue []int64) []int64 {
	var ret []int64
	x := reflect.ValueOf(v)
	switch x.Kind() {
	case reflect.Array, reflect.String, reflect.Slice:
		for i := 0; i < x.Len(); i++ {
			val := x.Index(i).Interface()
			if val == nil {
				return defaultValue
			}
			valInt := MustInt64(val, math.MaxInt64)
			if valInt == math.MaxInt64 {
				return defaultValue
			}
			ret = append(ret, valInt)
		}
		return ret
	default:
		return defaultValue
	}
}
