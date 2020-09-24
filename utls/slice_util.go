package utls

import (
	"strconv"
	"strings"
)

func StringArraySlice(array []string, offset int, length int) []string {
	count := len(array)
	if count == 0 || offset >= count {
		return nil
	}

	if length+offset > count {
		return array[offset:]
	}

	return array[offset : offset+length]
}

func Int64ArraySlice(array []int64, offset int, length int) []int64 {
	count := len(array)
	if count == 0 || offset >= count {
		return nil
	}

	if length+offset > count {
		return array[offset:]
	}

	return array[offset : offset+length]
}

//return subslice count and subslices
func CutStringSliceByStep(array []string, step int) (int, [][]string) {
	slicelen := len(array)
	if step >= slicelen {
		return 1, [][]string{array}
	}
	var ret [][]string
	groups := slicelen / step
	if groups*step < slicelen {
		groups++
	}
	for i := 0; i < groups; i++ {
		subslice := StringArraySlice(array, i*step, step)
		if len(subslice) > 0 {
			ret = append(ret, subslice)
		}
	}
	return groups, ret
}

func Int64ArrayToString(list []int64, sep string) string {

	if len(list) == 0 {
		return ""
	}
	ret := make([]string, 0, len(list))
	for _, l := range list {
		ret = append(ret, strconv.FormatInt(l, 10))
	}
	return strings.Join(ret, sep)
}

func StringInSlice(a []string, s string) bool {
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
