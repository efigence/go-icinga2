package icinga2

import (
	"encoding/json"
	"github.com/efigence/go-monitoring"
	"time"
)

type Icinga2APIResponse struct {
	Results []Icinga2APIObject
}

type Icinga2APIObject struct {
	Attrs json.RawMessage
	Type  string
	Name  string
}

type Icinga2APIHost struct {
	Name string `json:"name"`
	Active bool `json:"active"`
	State float32 `json:"last_state"`
	StateType float32 `json:"last_state_type"`
	LastCheck float64 `json:"last_check"`
	LastStateChange float64 `json:"last_state_change"`
	DowntimeDepth float32 `json:"downtime_depth"`
	Flapping bool `json:"flapping"`
	Acknowledgement float32 `json:"acknowledgement"`
	AcknowledgementExpiry float64 `json:"acknowledgement_expiry"`
}

type Icinga2APIService struct {
	Host string `json:"host_name"`
	Service string `json:"name"`
	Active bool `json:"active"`
	State float32 `json:"last_state"`
	StateType float32 `json:"last_state_type"`
	LastCheck float64 `json:"last_check"`
	LastStateChange float64 `json:"last_state_change"`
	DowntimeDepth float32 `json:"downtime_depth"`
	Flapping bool `json:"flapping"`
	Acknowledgement float32 `json:"acknowledgement"`
	AcknowledgementExpiry float64 `json:"acknowledgement_expiry"`

}

func (i *Icinga2APIResponse) GetHosts() (v []monitoring.Host) {
	for _, obj := range i.Results {
		if obj.Type != "Host" {	continue }
		var apiHost Icinga2APIHost
		err := json.Unmarshal(obj.Attrs,&apiHost)
		if err != nil {
			log.Printf("error unmarshalling host %s: %s | %s",obj.Name,err,string(obj.Attrs))
			continue
		}

		host := monitoring.Host{ Host: apiHost.Name }
		host.State =  uint8(apiHost.State) + 1
		host.Timestamp = unixTsToTs(apiHost.LastCheck)
		host.LastStateChange = unixTsToTs(apiHost.LastStateChange)
		host.Flapping = apiHost.Flapping
		if apiHost.StateType == 1.0 {
			host.StateHard = true
		}
		if apiHost.DowntimeDepth > 0 {
			host.Downtime = true
		}
		if apiHost.Acknowledgement > 0 {
			host.Acknowledged = true
		}
		v = append(v, host)
	}
	return v
}

func (i *Icinga2APIResponse) GetServices() (v []monitoring.Service) {
	for _, obj := range i.Results {
		if obj.Type != "Service" {	continue }
		var apiService Icinga2APIService
		err := json.Unmarshal(obj.Attrs,&apiService)
		if err != nil {
			log.Printf("error unmarshalling host %s: %s | %s",obj.Name,err,string(obj.Attrs))
			continue
		}

		service := monitoring.Service{
			Host: apiService.Host,
			Service: apiService.Service,
		}
		service.State =  uint8(apiService.State) + 1
		service.Timestamp = unixTsToTs(apiService.LastCheck)
		service.LastStateChange = unixTsToTs(apiService.LastStateChange)
		service.Flapping = apiService.Flapping
		if apiService.StateType == 1.0 {
			service.StateHard = true
		}
		if apiService.DowntimeDepth > 0 {
			service.Downtime = true
		}
		if apiService.Acknowledgement > 0 {
			service.Acknowledged = true
		}
		v = append(v, service)
	}
	return v
}


func unixTsToTs (t float64) time.Time {
	tsSec := int64(t)
	tsNs := int64((t - float64(tsSec)) * 1000000000)
	if tsNs < 0  {tsNs = 0}
	return time.Unix(tsSec,tsNs)
}
//func (i *Icinga2APIHostResponse)ToHost() *monitoring.Host {
//	var m monitoring.Host
//}

