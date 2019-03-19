package icinga2

import (
	"encoding/json"
	"fmt"
	"github.com/efigence/go-monitoring"
	"io/ioutil"
	"net/http"
	"net/url"
)

type API struct {
	URL *url.URL
	User string
	Pass string
}

func New(u,user, pass string) (ao *API, err error) {
	var a API
	a.URL, err = url.Parse(u)
	if err != nil {
		return ao, fmt.Errorf("error parsing url: %s", err)
	}
	a.User=user
	a.Pass=pass
	return &a, nil
}


func (a *API) GetHosts() (m []monitoring.Host, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", a.URL.String() + "/v1/objects/Hosts" , nil)
	if err != nil {return m, err}
	if len(a.User) > 0 {
		req.SetBasicAuth(a.User, a.Pass)
	}
	resp, err := client.Do(req)
	if err != nil {return m, err}
	var i Icinga2APIResponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {return m, err}
	err = json.Unmarshal(body,&i)
	if err != nil {
		return m, fmt.Errorf("error decoding json: %s | %s", err, string(body))
	}
	return i.GetHosts(), nil
}

func (a *API) GetServices() (m []monitoring.Service, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", a.URL.String() + "/v1/objects/Services" , nil)
	if err != nil {return m, err}
	if len(a.User) > 0 {
		req.SetBasicAuth(a.User, a.Pass)
	}
	resp, err := client.Do(req)
	if err != nil {return m, err}
	var i Icinga2APIResponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {return m, err}
	err = json.Unmarshal(body,&i)
	if err != nil {
		return m, fmt.Errorf("error decoding json: %s | %s", err, string(body))
	}
	return i.GetServices(), nil

}