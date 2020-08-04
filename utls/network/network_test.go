package network

import (
	"testing"

	"net/http"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetRealIP(t *testing.T) {
	Convey("Initialize a new http.Request object", t, func() {
		request, err := http.NewRequest("GET", "http://www.xiaomi.com/", nil)
		So(err, ShouldBeNil)
		Convey("Test X-Forwarded-For with single IP", func() {
			request.Header.Set("X-Forwarded-For", "120.52.112.1")
			ip, err := GetRealIP(request)
			So(err, ShouldBeNil)
			So(ip, ShouldEqual, "120.52.112.1")
		})
		Convey("Test X-Forwarded-For with IP lists", func() {
			request.Header.Set("X-Forwarded-For", "127.0.0.1, 120.52.112.1")
			ip, err := GetRealIP(request)
			So(err, ShouldBeNil)
			So(ip, ShouldEqual, "127.0.0.1")
		})
		Convey("Test RemoteAddr", func() {
			request.RemoteAddr = "127.0.0.1:8888"
			ip, err := GetRealIP(request)
			So(err, ShouldBeNil)
			So(ip, ShouldEqual, "127.0.0.1")
		})
		Convey("Test mixed X-Forwarded-For and RemoteAddr", func() {
			request.RemoteAddr = "192.168.1.1:9999"
			request.Header.Set("X-Forwarded-For", "120.52.112.1")
			ip, err := GetRealIP(request)
			So(err, ShouldBeNil)
			So(ip, ShouldEqual, "120.52.112.1")
		})
		Convey("Test invalid IP address (version 1)", func() {
			request.RemoteAddr = "invalid ip"
			ip, err := GetRealIP(request)
			So(err.Error(), ShouldContainSubstring, "cannot get real ip from request")
			So(ip, ShouldEqual, "")
		})
		Convey("Test invalid IP address (version 2)", func() {
			request.RemoteAddr = "127.0.0:fff"
			ip, err := GetRealIP(request)
			So(err.Error(), ShouldContainSubstring, "cannot get real ip from request")
			So(ip, ShouldEqual, "")
		})
	})
}

func TestIsLocalIP(t *testing.T) {
	Convey("Initialize a new http.Request object", t, func() {
		request, err := http.NewRequest("GET", "http://www.xiaomi.com/", nil)
		So(err, ShouldBeNil)
		Convey("Test Local IP 127.0.0.1", func() {
			request.RemoteAddr = "127.0.0.1:9999"
			isLocal, err := IsLocalIP(request)
			So(err, ShouldBeNil)
			So(isLocal, ShouldBeTrue)
		})
		Convey("Test Local IP 10.*", func() {
			request.RemoteAddr = "10.223.12.1:1111"
			isLocal, err := IsLocalIP(request)
			So(err, ShouldBeNil)
			So(isLocal, ShouldBeTrue)
		})
		Convey("Test Other IP", func() {
			request.RemoteAddr = "120.52.112.1:222"
			isLocal, err := IsLocalIP(request)
			So(err, ShouldBeNil)
			So(isLocal, ShouldBeFalse)
		})
		Convey("Test error messages", func() {
			isLocal, err := IsLocalIP(request)
			So(err.Error(), ShouldContainSubstring, "cannot get real ip from request")
			So(isLocal, ShouldBeFalse)
		})
	})
}
