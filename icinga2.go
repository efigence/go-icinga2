package icinga2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/efigence/go-monitoring"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var client = httpClient()

type API struct {
	URL  *url.URL
	User string
	Pass string
}

func New(apiURL, user, pass string) (ao *API, err error) {
	var a API
	a.URL, err = url.Parse(apiURL)
	if err != nil {
		return ao, fmt.Errorf("error parsing url: %s", err)
	}
	a.User = user
	a.Pass = pass
	return &a, nil
}

func httpClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 31,
	}

}

func (a *API) GetHosts() (m []monitoring.Host, err error) {
	return a.GetHostsByFilter("")
}

func (a *API) GetHostsByFilter(filter string) (m []monitoring.Host, err error) {
	client := httpClient()
	req, err := http.NewRequest("GET", a.URL.String()+"/v1/objects/Hosts", nil)
	if err != nil {
		return m, err
	}
	req.Header.Set("Accept", "application/json")
	if filter != "" {
		q := req.URL.Query()
		q.Add("filter", filter)
		req.URL.RawQuery = q.Encode()
	}
	if len(a.User) > 0 {
		req.SetBasicAuth(a.User, a.Pass)
	}
	resp, err := client.Do(req)
	if err != nil {
		return m, err
	}
	defer resp.Body.Close()
	var i Icinga2APIResponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(body, &i)
	if err != nil {
		return m, fmt.Errorf("error decoding json: %s | %s", err, string(body))
	}
	return i.GetHosts(), nil
}

func (a *API) GetServices() (m []monitoring.Service, err error) {
	return a.GetServicesByFilter("")
}

func (a *API) GetServicesByFilter(filter string) (m []monitoring.Service, err error) {
	client := httpClient()
	req, err := http.NewRequest("GET", a.URL.String()+"/v1/objects/Services", nil)
	if err != nil {
		return m, err
	}
	req.Header.Set("Accept", "application/json")
	if filter != "" {
		q := req.URL.Query()
		q.Add("filter", filter)
		req.URL.RawQuery = q.Encode()
	}
	if len(a.User) > 0 {
		req.SetBasicAuth(a.User, a.Pass)
	}
	resp, err := client.Do(req)
	if err != nil {
		return m, err
	}
	defer resp.Body.Close()
	var i Icinga2APIResponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(body, &i)
	if err != nil {
		return m, fmt.Errorf("error decoding json: %s | %s", err, string(body))
	}
	return i.GetServices(), nil
}

type Downtime struct {
	Flexible bool // Flexible downtime needs duration set. Nonflexible one needs start and stop time
	Start    time.Time
	End      time.Time
	Duration time.Duration // only for flexible downtime
	// all_services makes sense to be the default but is not for whatever reason
	// so we reverse it here
	NoAllServices bool
	Author        string
	Comment       string
}

func (d *Downtime) Validate() error {
	if d.Duration == 0 && d.Flexible {
		return fmt.Errorf("flexible downtime needs duration set")
	}
	if d.Start.IsZero() || d.End.IsZero() {
		return fmt.Errorf("downtime needs Start and Stop time")
	}
	return nil
}

type downtimeRequest struct {
	Type        string `json:"type"`
	Filter      string `json:"filter"`
	StartTime   int    `json:"start_time"`
	EndTime     int    `json:"end_time"`
	Duration    int    `json:"duration,omitempty"`
	Author      string `json:"author"`
	Comment     string `json:"comment"`
	AllServices bool   `json:"all_services,omitempty"`
}

// ScheduleHostDowntime schedules downtime on hostname by the match() filter of Icinga2 (so globs and such work)
func (a *API) ScheduleHostDowntime(host string, downtime Downtime) (downtimedHosts []string, err error) {
	return a.ScheduleHostDowntimeByFilter(`match("`+host+`", host.name)`, downtime)
}

// ScheduleHostDowntime schedules downtime on hostname by the icinga filters
func (a *API) ScheduleHostDowntimeByFilter(filter string, downtime Downtime) (downtimedHosts []string, err error) {
	err = downtime.Validate()
	if err != nil {
		return downtimedHosts, err
	}

	reqData := downtimeRequest{
		Type: "Host",
		// TODO wildcard protection
		Filter:      filter,
		StartTime:   int(downtime.Start.UTC().Unix()),
		EndTime:     int(downtime.End.UTC().Unix()),
		Author:      downtime.Author,
		Comment:     downtime.Comment,
		AllServices: !downtime.NoAllServices,
	}

	jsonData, err := json.Marshal(reqData)
	req, err := http.NewRequest("POST", a.URL.String()+"/v1/actions/schedule-downtime", bytes.NewBuffer(jsonData))
	if err != nil {
		return downtimedHosts, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if len(a.User) > 0 {
		req.SetBasicAuth(a.User, a.Pass)
	}
	resp, err := client.Do(req)
	if err != nil {
		return downtimedHosts, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return downtimedHosts, err
	}
	var i Icinga2StatusResponseOk
	err = json.Unmarshal(body, &i)
	if err != nil {
		return downtimedHosts, fmt.Errorf("error decoding json: %s | %s", err, string(body))
	}
	if len(i.Results) == 0 {
		var e Icinga2StatusPart
		err := json.Unmarshal(body, &e)
		if err != nil {
			return downtimedHosts, fmt.Errorf("error decoding json: %s | %s", err, string(body))
		}
		return downtimedHosts, fmt.Errorf("error while setting downtime for filter [%s]: zero hosts returned: [%s]", filter, string(body))
	}
	return i.GetDowntimeList(), nil
}
