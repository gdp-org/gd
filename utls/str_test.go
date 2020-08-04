package utls

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEscapeStr(t *testing.T) {
	Convey("test str escape", t, func() {
		nret := `{"code":0,"message":"ok","result":{"list":[{"id":"462193000","name":"监控开"},{"id":"463525564","name":"下班"},{"id":"468385450","name":"办公"},{"id":"468532940","name":"报警关"},{"id":"468539531","name":"监控关"},{"id":"468785637","name":"开关办公室照明"},{"id":"487728616","name":"烧水"},{"id":"488261579","name":"投影"}]}}`
		str := `{"code":0,"message":"ok","result":{"list":[{"id":"462193000","name":"监控开"},{"id":"463525564","name":"下班"},{"id":"468385450","name":"办公` + "\u0000" + `"},{"id":"468532940","name":"报警` + "\u0000" + `关"},{"id":"468539531","name":"监控关"},{"id":"468785637","name":"开关办公室照明"},{"id":"487728616","name":"烧水"},{"id":"488261579","name":"投影"}]}}`

		nstr := strings.Replace(str, "\u0000", "", -1)
		//nbts := bytes.Replace([]byte(str), []byte{0x00, 0x00}, []byte(""), -1)
		//nbts := bytes.Replace([]byte(str), []byte{0x00}, []byte(""), -1)
		//nstr := string(nbts)
		fmt.Println("escaped", nstr)
		So(nstr, ShouldEqual, nret)
	})
}

func TestSafeSprintf(t *testing.T) {
	Convey("test safeSprintf", t, func() {
		s := SafeSprintf("test %s %d", "1", 1)
		So(s, ShouldEqual, "test 1 1")

		s = SafeSprintf("test %s %d", "1", 1, 2)
		So(s, ShouldEqual, "test 1 1")

		//s = SafeSprintf("test %s %d", "1")
		//So(s, ShouldEqual, "test 1 %d")

		s = SafeSprintf("test %s %d %s", "1", 1, "1")
		So(s, ShouldEqual, "test 1 1 1")

		s = SafeSprintf("test %s %d %s %d %f", "1", 1, "!", 123, 1.1)
		So(s, ShouldEqual, "test 1 1 ! 123 1.100000")
	})
}
