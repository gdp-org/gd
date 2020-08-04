package utls

import (
	"testing"

	"reflect"

	. "github.com/smartystreets/goconvey/convey"
)

func TestArraySlice(t *testing.T) {
	Convey("testArraySlice", t, func() {
		a := make([]string, 0)
		offset := 0
		length := 1
		ret := StringArraySlice(a, offset, length)
		So(len(ret), ShouldEqual, 0)

		a = append(a, "1")
		ret = StringArraySlice(a, offset, length)
		So(len(ret), ShouldEqual, 1)

		length = 2
		ret = StringArraySlice(a, offset, length)
		So(len(ret), ShouldEqual, 1)

		offset = 1
		ret = StringArraySlice(a, offset, length)
		So(len(ret), ShouldEqual, 0)
	})
}

func TestCutStringSliceByStep(t *testing.T) {
	Convey("testArraySlice", t, func() {
		res := []string{"1", "2", "1", "2", "3", "4", "5", "6"}
		cnt, ret := CutStringSliceByStep(res, 1)
		So(cnt, ShouldEqual, 8)
		for k, v := range ret {
			So(len(v), ShouldEqual, 1)
			So(v[0], ShouldEqual, res[k])
		}
		cnt, ret = CutStringSliceByStep(res, 9)
		So(cnt, ShouldEqual, 1)
		So(reflect.DeepEqual(ret[0], res), ShouldBeTrue)
		cnt, ret = CutStringSliceByStep(res, 3)
		So(cnt, ShouldEqual, 3)
		for k, v := range ret {
			for x, y := range v {
				So(y, ShouldEqual, res[k*3+x])
			}
		}
	})
}
