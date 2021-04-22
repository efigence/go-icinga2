package icinga2

import (
	"fmt"
	"github.com/efigence/go-monitoring"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var TestUser = "testuser"
var TestPass = "testpass"

type testLogger struct {}
func (t testLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)

}

type testServ struct {
	*httptest.Server
	reqBody string
}

func testServer(t *testing.T, path string, filename string) *testServ {
	var tts = &testServ{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := ioutil.ReadFile("testdata/" + filename)
		assert.Equal(t,path,r.URL.Path)
		assert.Nil(t,err)
		b,_ := ioutil.ReadAll(r.Body)
		tts.reqBody = string(b)
		w.Write(f)
	}))
	tts.Server = ts
	return tts
}

func TestAPI_GetHosts(t *testing.T) {
	log = testLogger{}
	ts := testServer(t,"/v1/objects/Hosts","v1.objects.hosts.json")
	Api, err1 := New(ts.URL, TestUser, TestPass)
	hosts, err2 := Api.GetHosts()
	t.Run("parse input", func(t *testing.T) {
		assert.Nil(t, err1)
		assert.Nil(t, err2)

	})
	t.Run("length", func(t *testing.T) {
		assert.Len(t,hosts,7)
	})
	v := make(map[string]monitoring.Host,0)
	for _, h := range hosts {
		v[h.Host] = h
	}
	t.Run("host data", func(t *testing.T) {
		assert.Equal(t, monitoring.HostUp,int(v["t1-host1"].State))
		assert.Equal(t, "t1-host1", v["t1-host1"].DisplayName)
	})

}

func TestAPI_GetServices(t *testing.T) {
	log = testLogger{}
	ts := testServer(t,"/v1/objects/Services","v1.objects.services.json")
	Api, err1 := New(ts.URL, TestUser, TestPass)
	services, err2 := Api.GetServices()
	t.Run("parse input", func(t *testing.T) {
		assert.Nil(t, err1)
		assert.Nil(t, err2)

	})
	t.Run("length", func(t *testing.T) {
		assert.Len(t,services,7)
	})
	v := make(map[string]map[string]monitoring.Service,0)
	for _, s := range services {
		if _,ok := v[s.Host][s.Service]; !ok {
			v[s.Host] = make(map[string]monitoring.Service)
		}
		v[s.Host][s.Service] = s
	}
	t.Run("service data", func(t *testing.T) {
		assert.True(t, v["t1-host1"]["ELASTICSEARCH"].Flapping)
		assert.False(t,v["t1-host1"]["ELASTICSEARCH"].Acknowledged)
		assert.Equal(t,monitoring.StatusOk,int(v["t1-host1"]["ELASTICSEARCH"].State))
	})
}

func TestAPI_ScheduleHostDowntime(t *testing.T) {
	log = testLogger{}
	ts := testServer(t, "/v1/actions/schedule-downtime", "v1.actions.schedule-downtime.json")
	Api, err1 := New(ts.URL, TestUser, TestPass)
	hosts, err2 := Api.ScheduleHostDowntime("t1-host1", Downtime{
		Flexible:      false,
		Start:         time.Now(),
		End:           time.Now().Add(time.Hour),
		Duration:      0,
		NoAllServices: false,
		Author: t.Name(),
		Comment: "c:" + t.Name(),
	})
	t.Run("parse input", func(t *testing.T) {
		assert.Nil(t, err1)
		assert.Nil(t, err2)
	})
	t.Run("request json", func(tt *testing.T) {
		assert.Contains(tt, ts.reqBody,`"filter":"match(\"t1-host1\", host.name)"`)
		assert.Contains(tt, ts.reqBody,`"author":"` + t.Name() + `"`)
		assert.Contains(tt, ts.reqBody,`"comment":"c:` + t.Name() + `"`)
	})
	t.Run("host count", func(t *testing.T) {
		assert.Len(t,hosts,2)
	})

}

func TestAPI_ScheduleHostDowntime_NoHost(t *testing.T) {
	log = testLogger{}
	ts := testServer(t, "/v1/actions/schedule-downtime", "error.no-objects-found.json")
	Api, err1 := New(ts.URL, TestUser, TestPass)
	_, err2 := Api.ScheduleHostDowntime("t1-host1", Downtime{
		Flexible:      false,
		Start:         time.Now(),
		End:           time.Now().Add(time.Hour),
		Duration:      0,
		NoAllServices: false,

	})
	t.Run("parse input", func(t *testing.T) {
		assert.Nil(t, err1)
		assert.Error(t, err2)
	})
}