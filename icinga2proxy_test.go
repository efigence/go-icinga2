package icinga2

import (
	"github.com/efigence/go-monitoring"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)



func TestProxy_GetHosts_Single(t *testing.T) {
	log = testLogger{}
	ts := testServer(t, "/v1/objects/Hosts", "v1.objects.hosts.json")
	Api, err1 := NewProxy(map[string]Icinga2ServerConfig{
		"s1" : {ts.URL, TestUser, TestPass},
	})
	hosts, err2 := Api.GetHosts()
	t.Run("parse input", func(t *testing.T) {
		assert.Nil(t, err1)
		assert.Nil(t, err2)

	})
	t.Run("length", func(t *testing.T) {
		assert.Len(t, hosts, 7)
	})
	v := make(map[string]monitoring.Host, 0)
	for _, h := range hosts {
		v[h.Host] = h
	}
	t.Run("host data", func(t *testing.T) {
		assert.Equal(t, monitoring.HostUp, int(v["t1-host1"].State))
		assert.Equal(t, "t1-host1", v["t1-host1"].DisplayName)
	})
}

func TestProxy_GetHosts_Dedup(t *testing.T) {
	log = testLogger{}
	ts := testServer(t, "/v1/objects/Hosts", "v1.objects.hosts.json")
	ts2 := testServer(t, "/v1/objects/Hosts", "v1.objects.hosts_dedup.json")
	Api, err1 := NewProxy(map[string]Icinga2ServerConfig{
		"s1" : {ts.URL, TestUser, TestPass},
		"s2" : {ts2.URL, TestUser, TestPass},
	})
	hosts, err2 := Api.GetHosts()
	t.Run("parse input", func(t *testing.T) {
		assert.Nil(t, err1)
		assert.Nil(t, err2)

	})
	t.Run("length", func(t *testing.T) {
		assert.Len(t, hosts, 14)
	})
	v := make(map[string]monitoring.Host, 0)
	for _, h := range hosts {
		v[h.Host] = h
	}
	t.Run("duped host data", func(t *testing.T) {
		assert.Equal(t, monitoring.HostUp, int(v["t1-host1_s1"].State))
		assert.Equal(t, "t1-host1", v["t1-host1_s1"].DisplayName,"dedup should not change display name")
	})
	t.Run("single host data", func(t *testing.T) {
		assert.Equal(t, monitoring.HostUp, int(v["t2-lb1"].State))
		assert.Equal(t, "t2-lb1", v["t2-lb1"].DisplayName,)
	})
}


func TestProxy_GetServices_Single(t *testing.T) {
	log = testLogger{}
	ts := testServer(t, "/v1/objects/Services", "v1.objects.services.json")
	Api, err1 := NewProxy(map[string]Icinga2ServerConfig{
		"s1" : {ts.URL, TestUser, TestPass},
	})
	services, err2 := Api.GetServices()
	t.Run("parse input", func(t *testing.T) {
		assert.Nil(t, err1)
		assert.Nil(t, err2)

	})
	t.Run("length", func(t *testing.T) {
		assert.Len(t, services, 7, "should not deduplicate")
	})
	v := make(map[string]map[string]monitoring.Service, 0)
	for _, s := range services {
		if _, ok := v[s.Host][s.Service]; !ok {
			v[s.Host] = make(map[string]monitoring.Service)
		}
		v[s.Host][s.Service] = s
	}
	t.Run("service data", func(t *testing.T) {
		assert.True(t, v["t1-host1"]["ELASTICSEARCH"].Flapping)
		assert.False(t, v["t1-host1"]["ELASTICSEARCH"].Acknowledged)
		assert.Equal(t, monitoring.StatusOk, int(v["t1-host1"]["ELASTICSEARCH"].State))
	})
}



func TestProxy_GetServices_Dedup(t *testing.T) {
	log = testLogger{}
	ts := testServer(t, "/v1/objects/Services", "v1.objects.services.json")
	ts2 := testServer(t, "/v1/objects/Services", "v1.objects.services_dedup.json")
	Api, err1 := NewProxy(map[string]Icinga2ServerConfig{
		"s1" : {ts.URL, TestUser, TestPass},
		"s2" : {ts2.URL, TestUser, TestPass},
	})
	services, err2 := Api.GetServices()
	t.Run("parse input", func(t *testing.T) {
		assert.Nil(t, err1)
		assert.Nil(t, err2)

	})
	t.Run("length", func(t *testing.T) {
		assert.Len(t, services, 14, "should not deduplicate")
	})
	v := make(map[string]map[string]monitoring.Service, 0)
	for _, s := range services {
		if _, ok := v[s.Host][s.Service]; !ok {
			v[s.Host] = make(map[string]monitoring.Service)
		}
		v[s.Host][s.Service] = s
	}
	t.Run("DUPed host service data", func(t *testing.T) {
		assert.True(t, v["t1-host1_s1"]["ELASTICSEARCH"].Flapping)
		assert.False(t, v["t1-host1_s1"]["ELASTICSEARCH"].Acknowledged)
		assert.Equal(t, monitoring.StatusOk, int(v["t1-host1_s1"]["ELASTICSEARCH"].State))
	})

	t.Run("single service data", func(t *testing.T) {
		assert.True(t, v["t2-lb1"]["CERTIFICATE-APP.EXAMPLE.COM"].Flapping)
		assert.False(t, v["t2-lb1"]["CERTIFICATE-APP.EXAMPLE.COM"].Acknowledged)
		assert.Equal(t, monitoring.StatusOk, int(v["t2-lb1"]["CERTIFICATE-APP.EXAMPLE.COM"].State))
	})
}

func TestProxy_ScheduleHostDowntime(t *testing.T) {
	log = testLogger{}
	ts := testServer(t, "/v1/actions/schedule-downtime", "v1.actions.schedule-downtime.json")
	Api, err1 := NewProxy(map[string]Icinga2ServerConfig{
		"s1" : {ts.URL, TestUser, TestPass},
		"s2" : {ts.URL, TestUser, TestPass},
	})
	hosts, err2 := Api.ScheduleHostDowntime("t1-host1", Downtime{
		Flexible:      false,
		Start:         time.Now(),
		End:           time.Now().Add(time.Hour),
		Duration:      0,
		NoAllServices: false,
		Author:        t.Name(),
		Comment:       "c:" + t.Name(),
	})
	t.Run("parse input", func(t *testing.T) {
		assert.Nil(t, err1)
		assert.Nil(t, err2)
	})
	t.Run("request json", func(tt *testing.T) {
		assert.Contains(tt, ts.reqBody, `"filter":"match(\"t1-host1\", host.name)"`)
		assert.Contains(tt, ts.reqBody, `"author":"`+t.Name()+`"`)
		assert.Contains(tt, ts.reqBody, `"comment":"c:`+t.Name()+`"`)
	})
	t.Run("host count", func(t *testing.T) {
		assert.Len(t, hosts, 2)
	})

}

func TestProxy_ScheduleHostDowntime_NoHost(t *testing.T) {
	log = testLogger{}
	ts := testServer(t, "/v1/actions/schedule-downtime", "error.no-objects-found.json")
	Api, err1 := NewProxy(map[string]Icinga2ServerConfig{
		"s1" : {ts.URL, TestUser, TestPass},
	})
	_, err2 := Api.ScheduleHostDowntime("t1-host1", Downtime{
		Flexible:      false,
		Start:         time.Now(),
		End:           time.Now().Add(time.Hour),
		Duration:      0,
		NoAllServices: false,
		Author: "testAuthor",
		Comment: "testComment",
	})
	t.Run("parse input", func(t *testing.T) {
		assert.Nil(t, err1)
		assert.Error(t, err2)
		assert.Contains(t,ts.reqBody,`"author":"testAuthor"`)
		assert.Contains(t,ts.reqBody,`"comment":"testComment"`)
		assert.Contains(t,ts.reqBody,`"filter":"match(\"t1-host1\", host.name)"`)
	})
}
