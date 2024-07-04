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

	for _, r := range responseServers {

		info, err := stun.ParseURI(r.URLs)

		if err != nil {
			return nil, err
		}

		if ((info.Scheme == stun.SchemeTypeTURN || info.Scheme == stun.SchemeTypeTURNS) && !d.Config.TurnEnabled) || ((info.Scheme == stun.SchemeTypeSTUN || info.Scheme == stun.SchemeTypeSTUNS) && !d.Config.StunEnabled) {
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
	}
	return iceServers, nil
}
