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
	Name string `json:"__name"`
	Active bool `json:"active"`
	State float32 `json:"last_state"`
	StateType float32 `json:"last_state_type"`
	LastCheck float64 `json:"last_check"`
	LastStateChange float64 `json:"last_state_change"`

}

func (i *Icinga2APIResponse)GetHosts() (v []monitoring.Host) {
	for _, obj := range i.Results {
		if obj.Type != "Host" {	continue }
		var apiHost Icinga2APIHost
		err := json.Unmarshal(obj.Attrs,&apiHost)
		if err != nil {
			log.Printf("error unmarshalling host %s: %s | %s",obj.Name,err,string(obj.Attrs))
			continue
		}

		host := monitoring.Host {
			Host: apiHost.Name,
			State: uint8(apiHost.State),
			Timestamp: unixTsToTs(apiHost.LastCheck),
			LastStateChange: unixTsToTs(apiHost.LastStateChange),
		}
		v = append(v, host)
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

