package icinga2

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

var TestURL = os.Getenv("ICINGA2_URL")
var TestUser = os.Getenv("ICINGA2_USER")
var TestPass = os.Getenv("ICINGA2_PASS")

type testLogger struct {}
func (t testLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)

}

func TestAPI_GetHosts(t *testing.T) {
	log = testLogger{}
	Api, err1 := New(TestURL, TestUser, TestPass)
	_, err2 := Api.GetHosts()
	Convey("No separators",t,func() {
		So(err1,ShouldBeNil)
		So(err2,ShouldBeNil)
	})
}

func TestAPI_GetServices(t *testing.T) {
	log = testLogger{}
	Api, err1 := New(TestURL, TestUser, TestPass)
	_, err2 := Api.GetServices()
	Convey("No separators",t,func() {
		So(err1,ShouldBeNil)
		So(err2,ShouldBeNil)
	})
}