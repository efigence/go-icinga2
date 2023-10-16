package icinga2

import (
	"encoding/json"
	"github.com/efigence/go-monitoring"
	"strings"
	"time"
)

type Icinga2APIResponse struct {
	Results []Icinga2APIObject
}

type Icinga2StatusResponseOk struct {
	Results []Icinga2StatusPart
}

type Icinga2StatusPart struct {
	Code   float64 `json:"code"`  // yes, api returns floats as status code
	Error  float64 `json:"error"` // yes, api returns floats as error code
	Status string  `json:"status"`
	Name   string  `json:"name"`
}

type Icinga2APIObject struct {
	Attrs json.RawMessage
	Type  string
	Name  string
}

type Icinga2APIHost struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Active      bool    `json:"active"`
	State       float32 `json:"state"`
	StateType   float32 `json:"state_type"`
	// TODO find the difference between the two
	LastState             float32               `json:"last_state"`
	LastStateType         float32               `json:"last_state_type"`
	LastCheck             float64               `json:"last_check"`
	LastStateChange       float64               `json:"last_state_change"`
	LastHardStateChange   float64               `json:"last_hard_state_change"`
	DowntimeDepth         float32               `json:"downtime_depth"`
	Flapping              bool                  `json:"flapping"`
	Acknowledgement       float32               `json:"acknowledgement"`
	AcknowledgementExpiry float64               `json:"acknowledgement_expiry"`
	ActionURL             string                `json:"action_url"`
	NotesURL              string                `json:"notes_url"`
	CheckResult           Icinga2APICheckResult `json:"last_check_result"`
}

type Icinga2APIService struct {
	Host        string  `json:"host_name"`
	Service     string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Active      bool    `json:"active"`
	State       float32 `json:"state"`
	StateType   float32 `json:"state_type"`
	// TODO find the difference between the two
	LastState             float32               `json:"last_state"`
	LastStateType         float32               `json:"last_state_type"`
	LastCheck             float64               `json:"last_check"`
	LastStateChange       float64               `json:"last_state_change"`
	LastHardStateChange   float64               `json:"last_hard_state_change"`
	DowntimeDepth         float32               `json:"downtime_depth"`
	Flapping              bool                  `json:"flapping"`
	Acknowledgement       float32               `json:"acknowledgement"`
	AcknowledgementExpiry float64               `json:"acknowledgement_expiry"`
	ActionURL             string                `json:"action_url"`
	NotesURL              string                `json:"notes_url"`
	CheckResult           Icinga2APICheckResult `json:"last_check_result"`
}

type Icinga2APICheckResult struct {
	CheckFrom string `json:"check_source"`
	Message   string `json:"output"`
}

func (i *Icinga2APIResponse) GetHosts() (v []monitoring.Host) {
	for _, obj := range i.Results {
		if obj.Type != "Host" {
			continue
		}
		var apiHost Icinga2APIHost
		err := json.Unmarshal(obj.Attrs, &apiHost)
		if err != nil {
			log.Printf("error unmarshalling host %s: %s | %s", obj.Name, err, string(obj.Attrs))
			continue
		}

		host := monitoring.Host{Host: apiHost.Name}
		host.State = uint8(apiHost.State) + 1
		host.Timestamp = unixTsToTs(apiHost.LastCheck)
		host.LastStateChange = unixTsToTs(apiHost.LastStateChange)
		host.LastHardStateChange = unixTsToTs(apiHost.LastHardStateChange)
		host.Flapping = apiHost.Flapping
		host.CheckMessage = apiHost.CheckResult.Message
		host.DisplayName = apiHost.DisplayName
		if apiHost.StateType == 1.0 {
			host.StateHard = true
		}
		if apiHost.DowntimeDepth > 0 {
			host.Downtime = true
		}
		if apiHost.Acknowledgement > 0 {
			host.Acknowledged = true
		}
		if len(apiHost.ActionURL) > 0 {
			host.URL = apiHost.ActionURL
		} else if len(apiHost.NotesURL) > 0 {
			host.URL = apiHost.NotesURL
		}
		v = append(v, host)
	}
	return v
}

func (i *Icinga2APIResponse) GetServices() (v []monitoring.Service) {
	for _, obj := range i.Results {
		if obj.Type != "Service" {
			continue
		}
		var apiService Icinga2APIService
		err := json.Unmarshal(obj.Attrs, &apiService)
		if err != nil {
			log.Printf("error unmarshalling host %s: %s | %s", obj.Name, err, string(obj.Attrs))
			continue
		}

		service := monitoring.Service{
			Host:    apiService.Host,
			Service: apiService.Service,
		}
		service.State = uint8(apiService.State) + 1
		service.Timestamp = unixTsToTs(apiService.LastCheck)
		service.LastStateChange = unixTsToTs(apiService.LastStateChange)
		service.Flapping = apiService.Flapping
		service.CheckMessage = apiService.CheckResult.Message
		service.DisplayName = apiService.DisplayName
		if apiService.StateType == 1.0 {
			service.StateHard = true
		}
		if apiService.DowntimeDepth > 0 {
			service.Downtime = true
		}
		if apiService.Acknowledgement > 0 {
			service.Acknowledged = true
		}
		if len(apiService.ActionURL) > 0 {
			service.URL = apiService.ActionURL
		} else if len(apiService.NotesURL) > 0 {
			service.URL = apiService.NotesURL
		}

		v = append(v, service)
	}
	return v
}

func (i *Icinga2StatusResponseOk) GetDowntimeList() []string {
	objects := make([]string, 0)
	for _, obj := range i.Results {
		if int(obj.Code) == 200 {
			parts := strings.SplitN(obj.Name, "!", 2)
			objects = append(objects, parts[0])
		} else {
			// TODO handle somehow ?
		}
	}
	return objects
}

func unixTsToTs(t float64) time.Time {
	tsSec := int64(t)
	tsNs := int64((t - float64(tsSec)) * 1000000000)
	if tsNs < 0 {
		tsNs = 0
	}
	return time.Unix(tsSec, tsNs)
}

//func (i *Icinga2APIHostResponse)ToHost() *monitoring.Host {
//	var m monitoring.Host
//}
