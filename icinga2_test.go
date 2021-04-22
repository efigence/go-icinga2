package icinga2

import (
	"fmt"
	"github.com/efigence/go-monitoring"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var TestUser = "testuser"
var TestPass = "testpass"

type testLogger struct {}
func (t testLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)

}

func testServer(t *testing.T, path string, filename string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := ioutil.ReadFile("testdata/" + filename)
		assert.Equal(t,path,r.URL.Path)
		assert.Nil(t,err)
		w.Write(f)
	}))
	return ts
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