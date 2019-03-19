package icinga2

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

var TestURL = os.Getenv("ICINGA2_URL")
var TestUser = os.Getenv("ICINGA2_USER")
var TestPass = os.Getenv("ICINGA2_PASS")

func TestAPI_GetHosts(t *testing.T) {
	Api, err1 := New(TestURL, TestUser, TestPass)
	_, err2 := Api.GetHosts()
	Convey("No separators",t,func() {
		So(err1,ShouldBeNil)
		So(err2,ShouldBeNil)
	})
}