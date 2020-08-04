package utls

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

func Str2Bts(str string) []byte {
	s := *(*reflect.StringHeader)(unsafe.Pointer(&str))
	b := &reflect.SliceHeader{Data: s.Data, Len: s.Len, Cap: s.Len}
	return *(*[]byte)(unsafe.Pointer(b))
}

func Bts2Str(bts []byte) string {
	b := *(*reflect.SliceHeader)(unsafe.Pointer(&bts))
	s := &reflect.StringHeader{Data: b.Data, Len: b.Len}
	return *(*string)(unsafe.Pointer(s))
}

//TODO://解决参数类型不对的问题,以及％.2f类似问题
func SafeSprintf(format string, args ...interface{}) string {
	strCount := strings.Count(format, "%s")
	intCount := strings.Count(format, "%d")
	floatCount := strings.Count(format, "%f")
	count := strCount + intCount + floatCount
	argsLen := len(args)
	if count > argsLen {
		count = argsLen
	}
	return fmt.Sprintf(format, args[0:count]...)
}
