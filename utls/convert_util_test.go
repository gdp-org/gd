package utls

import (
	"math"
	"testing"

	"encoding/json"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConvertUtil_MustString(t *testing.T) {
	Convey("muststring", t, func() {
		res := MustString("123", "1")
		So(res, ShouldEqual, "123")

		res = MustString([]byte("123"), "1")
		So(res, ShouldEqual, "123")

		res = MustString(int(123), "1")
		So(res, ShouldEqual, "123")

		res = MustString(int64(123), "1")
		So(res, ShouldEqual, "123")

		res = MustString(uint64(123), "1")
		So(res, ShouldEqual, "123")

		res = MustString(int32(123), "1")
		So(res, ShouldEqual, "123")

		res = MustString(uint32(123), "1")
		So(res, ShouldEqual, "123")

		res = MustString([]string{}, "1")
		So(res, ShouldEqual, "1")
	})
}

func TestConvertUtil_MustInt64(t *testing.T) {
	Convey("mustint64", t, func() {

		res := MustInt64(nil, 1)
		So(res, ShouldEqual, 1)

		res = MustInt64("123", 1)
		So(res, ShouldEqual, 123)

		res = MustInt64("aaa123", 1)
		So(res, ShouldEqual, 1)

		res = MustInt64([]byte("123"), 1)
		So(res, ShouldEqual, 123)

		res = MustInt64(int(123), 1)
		So(res, ShouldEqual, 123)

		res = MustInt64(int64(123), 1)
		So(res, ShouldEqual, 123)

		res = MustInt64(uint64(123), 1)
		So(res, ShouldEqual, 123)

		res = MustInt64(uint64(math.MaxInt64)+5, 1)
		So(res, ShouldEqual, 1)

		res = MustInt64(int32(math.MaxInt32), 1)
		So(res, ShouldEqual, math.MaxInt32)

		res = MustInt64(uint32(math.MaxUint32), 1)
		So(res, ShouldEqual, math.MaxUint32)

		res = MustInt64([]string{}, 1)
		So(res, ShouldEqual, 1)

		res = MustInt64(json.Number("123"), 1)
		So(res, ShouldEqual, 123)

	})
}

func TestConvertUtil_MustInt64Array(t *testing.T) {
	Convey("test_must_int64", t, func() {
		testTable := []*struct {
			Input  interface{}
			Output []int64
		}{
			{
				Input:  []int{1, 2, 3, 4},
				Output: []int64{1, 2, 3, 4},
			},
			{
				Input:  []interface{}{"1", 2, "3", 4},
				Output: []int64{1, 2, 3, 4},
			},
			{
				Input:  []interface{}{"1", map[string]string{}, "3", 4},
				Output: []int64{},
			},
			{
				Input:  []interface{}{"1", json.Number("2"), "3", 4},
				Output: []int64{1, 2, 3, 4},
			},
		}

		for _, ca := range testTable {
			res := MustInt64Array(ca.Input, []int64{})
			So(len(res), ShouldEqual, len(ca.Output))
			for i, r := range res {
				So(r, ShouldEqual, ca.Output[i])
			}
		}
	})
}

func TestConvertUtil_MustStringArray(t *testing.T) {
	Convey("test_must_string", t, func() {
		testTable := []*struct {
			Input  interface{}
			Output []string
		}{
			{
				Input:  []int{1, 2, 3, 4},
				Output: []string{"1", "2", "3", "4"},
			},
			{
				Input:  []interface{}{"1", 2, "3", 4},
				Output: []string{"1", "2", "3", "4"},
			},
			{
				Input:  []interface{}{"1", map[string]string{}, "3", 4},
				Output: []string{},
			},
			{
				Input:  []interface{}{"1", json.Number("2"), "3", 4},
				Output: []string{"1", "2", "3", "4"},
			},
		}

		for _, ca := range testTable {
			res := MustStringArray(ca.Input, []string{})
			So(len(res), ShouldEqual, len(ca.Output))
			for i, r := range res {
				So(r, ShouldEqual, ca.Output[i])
			}
		}
	})
}
