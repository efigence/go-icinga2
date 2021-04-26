package icinga2

import (
	"fmt"
	"github.com/efigence/go-monitoring"
	"sync"
)

type Icinga2ServerConfig struct {
	ServerURL string `yaml:"server_url"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

type Proxy struct {
	servers map[string]*API
	conflictResolver func(icinga2ServerName string, hostName string) (newName string)
}

// NewProxy creates Icinga2 proxy that will merge data from all servers
func NewProxy(servers map[string]Icinga2ServerConfig,) (*Proxy, error) {
	p := &Proxy{
		servers: make(map[string]*API,0),
		conflictResolver: func(icinga2ServerName string, hostName string) (newName string) {
			return fmt.Sprintf("%s_%s",hostName,icinga2ServerName)
		},
	}
	for k, s := range servers {
		server, err := New(s.ServerURL,s.User,s.Pass)
		if err != nil {
			return nil, fmt.Errorf("error configuring icinga2 instance at %s:%s", s.ServerURL, err)
		}
		p.servers[k] = server
	}
	return p,nil
}

func sliceToMapHost(s []monitoring.Host) (map[string]monitoring.Host) {
	m := make(map[string]monitoring.Host,len(s))
	for _, a := range s {
		m[a.Host] = a
	}
	return m
}


func (a *Proxy)SetConflictResolver(f func(icinga2ServerName string, hostName string) (newName string)) {

}

func (a *Proxy) GetHosts() (m []monitoring.Host, err error) {
return a.GetHostsByFilter("")

}
func (a *Proxy) GetHostsByFilter(filter string) (m []monitoring.Host, err error) {
	var wg sync.WaitGroup
	res := make( map[string]map[string]monitoring.Host)
	errs := make( map[string]error)
	for k, v := range a.servers {
		wg.Add(1)
		func(k string) {
			var z []monitoring.Host
			z, errs[k]  = v.GetHostsByFilter(filter)
			res[k]= sliceToMapHost(z)
			wg.Done()
		}(k)
	}
	collisionMap := make(map[string]int8,0)
	out := make([]monitoring.Host,0)
	wg.Wait()
	for _, hosts := range res {
		for host, _ := range hosts {
			collisionMap[host]++
		}
	}
	for icingaServer, hosts := range res {
		for host, check := range hosts {
			if collisionMap[host] > 1 {
				check.Host = a.conflictResolver(icingaServer,host)
			}
			out = append(out,check)
		}
	}
	errOut := true
	for _, err := range errs {
		if err == nil {
			errOut = false
		}
	}
	// TODO figure out how to signal that. Err handler for logging ?
	if errOut {
		return out, fmt.Errorf("error: [%+v]", errs)
	} else {
		return out, nil
	}


}
func (a *Proxy) GetServices() (m []monitoring.Service, err error) {
	return a.GetServicesByFilter("")
}
func (a *Proxy) GetServicesByFilter(filter string) (m []monitoring.Service, err error) {
	var wg sync.WaitGroup
	res := make( map[string][]monitoring.Service)
	errs := make( map[string]error)
	for k, v := range a.servers {
		wg.Add(1)
		func(k string) {
			res[k], errs[k]  = v.GetServicesByFilter(filter)
			wg.Done()
		}(k)
	}
	partialCollisionMap := make(map[string]map[string]bool)
	collisionMap := make(map[string]int8,0)
	out := make([]monitoring.Service,0)
	wg.Wait()
	for h, services := range res {
		partialCollisionMap[h] = make(map[string]bool,0)
		for _, serviceData := range services {
			partialCollisionMap[h][serviceData.Host]=true
		}
	}
	for _, services := range partialCollisionMap {
		for host, _ :=  range services {
			collisionMap[host]++
		}
	}
	for icingaServer, services := range res {
		for _, check := range services {
			if collisionMap[check.Host] > 1 {
				check.Host=a.conflictResolver(icingaServer,check.Host)
			}
			out = append(out,check)

		}
	}
	errOut := true
	for _, err := range errs {
		if err == nil {
			errOut = false
		}
	}
	// TODO figure out how to signal that. Err handler for logging ?
	if errOut {
		return out, fmt.Errorf("error: [%+v]", errs)
	} else {
		return out, nil
	}



}
func (a *Proxy) ScheduleHostDowntime(host string,downtime Downtime) (downtimedHosts []string, err error) {
	return a.ScheduleHostDowntimeByFilter(`match("` + host + `", host.name)`,downtime)
}
func (a *Proxy) ScheduleHostDowntimeByFilter(filter string, downtime Downtime) (downtimedHosts []string, err error) {
	res := make( map[string][]string)
	errs := make( map[string]error)
	var wg sync.WaitGroup
	for k, v := range a.servers {
		wg.Add(1)
		func(k string) {
			res[k], errs[k]  = v.ScheduleHostDowntimeByFilter(filter,downtime)
			wg.Done()
		}(k)
	}
	wg.Wait()
	downtimeMap := make(map[string]bool)
	for _, hosts := range res {
		for _, host := range hosts{
			downtimeMap[host]=true
		}
	}
	out := make([]string,0)
	for h, _ := range downtimeMap {
		out = append(out, h)
	}

	errOut := true
	for _, err := range errs {
		if err == nil {
			errOut = false
		}
	}
	// TODO figure out how to signal that. Err handler for logging ?
	if errOut {
		return out, fmt.Errorf("error: [%+v]", errs)
	} else {
		return out, nil
	}


}
