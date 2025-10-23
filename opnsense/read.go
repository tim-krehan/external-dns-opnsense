package opnsense

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

func (override *OpnSenseHostOverride) Read(api *OpnSenseApi, uuid string) error {
	endpoint := fmt.Sprintf("/unbound/settings/get_host_override/%s", uuid)
	var reponseHost struct {
		Host struct {
			Enabled  string `json:"enabled"`
			HostName string `json:"hostname"`
			Domain   string `json:"domain"`
			RR       struct {
				A struct {
					Value    string `json:"value"`
					Selected string `json:"selected"`
				} `json:"A"`
				AAAA struct {
					Value    string `json:"value"`
					Selected string `json:"selected"`
				} `json:"AAAA"`
				MX struct {
					Value    string `json:"value"`
					Selected string `json:"selected"`
				} `json:"MX"`
				TXT struct {
					Value    string `json:"value"`
					Selected string `json:"selected"`
				} `json:"TXT"`
			} `json:"rr"`
			MxPrio      string `json:"mxprio"`
			Mx          string `json:"mx"`
			TTL         string `json:"ttl"`
			Server      string `json:"server"`
			TxtData     string `json:"txtdata"`
			Description string `json:"description"`
		} `json:"host"`
	}
	resp, err := api.ApiRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&reponseHost)
	if err != nil {
		return err
	}

	values := reflect.ValueOf(reponseHost.Host.RR)
	recType := ""
	for i := 0; i < values.NumField(); i++ {
		field := values.Field(i)
		selected := field.FieldByName("Selected").String()
		if selected == "1" {
			recType = values.Type().Field(i).Name
			break
		}
	}
	override.Enabled = reponseHost.Host.Enabled
	override.HostName = reponseHost.Host.HostName
	override.Domain = reponseHost.Host.Domain
	override.Type = recType
	override.MxPrio = reponseHost.Host.MxPrio
	override.Mx = reponseHost.Host.Mx
	override.TTL = reponseHost.Host.TTL
	override.Server = reponseHost.Host.Server
	override.TxtData = reponseHost.Host.TxtData
	override.Description = reponseHost.Host.Description

	return nil
}
