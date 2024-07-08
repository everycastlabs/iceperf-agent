package metered

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/nimbleape/iceperf-agent/config"
	"github.com/pion/stun/v2"
	"github.com/pion/webrtc/v4"
)

type Driver struct {
	Config *config.ICEConfig
}

type MeteredIceServers struct {
	URLs       string `json:"urls,omitempty"`
	Username   string `json:"username,omitempty"`
	Credential string `json:"credential,omitempty"`
}

// func (d Driver) Measure(measurementName string) error {
// 	return nil
// }

func (d *Driver) GetIceServers() (iceServers []webrtc.ICEServer, err error) {
	res, err := http.Get(d.Config.RequestUrl + "?apiKey=" + d.Config.ApiKey)
	if err != nil {
		// log.WithFields(log.Fields{
		// 	"error": err,
		// }).Error("Error making http request")
		return nil, err
	}

	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		// log.WithFields(log.Fields{
		// 	"error": err,
		// }).Error("Error reading http response")
		return nil, err
	}

	var responseServers []MeteredIceServers
	json.Unmarshal([]byte(responseData), &responseServers)

	gotTransports := make(map[string]bool)

	gotTransports[stun.SchemeTypeSTUN.String()+stun.ProtoTypeUDP.String()] = false
	gotTransports[stun.SchemeTypeSTUN.String()+stun.ProtoTypeTCP.String()] = false
	gotTransports[stun.SchemeTypeTURN.String()+stun.ProtoTypeUDP.String()] = false
	gotTransports[stun.SchemeTypeTURN.String()+stun.ProtoTypeTCP.String()] = false
	gotTransports[stun.SchemeTypeTURNS.String()+stun.ProtoTypeTCP.String()] = false

	for _, r := range responseServers {

		info, err := stun.ParseURI(r.URLs)

		if err != nil {
			return nil, err
		}

		if ((info.Scheme == stun.SchemeTypeTURN || info.Scheme == stun.SchemeTypeTURNS) && !d.Config.TurnEnabled) || ((info.Scheme == stun.SchemeTypeSTUN || info.Scheme == stun.SchemeTypeSTUNS) && !d.Config.StunEnabled) {
			continue
		}

		if gotTransports[info.Scheme.String()+info.Proto.String()] {
			//we don't want to test all the special ports right now
			continue
		}

		s := webrtc.ICEServer{
			URLs: []string{r.URLs},
		}

		if r.Username != "" {
			s.Username = r.Username
		}
		if r.Credential != "" {
			s.Credential = r.Credential
		}
		iceServers = append(iceServers, s)
		gotTransports[info.Scheme.String()+info.Proto.String()] = true
	}
	return iceServers, nil
}
