package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/nimbleape/iceperf-agent/config"
	"github.com/pion/webrtc/v4"
)

type Driver struct {
	Config *config.ICEConfig
}

type ApiIceServer struct {
	URL        string `json:"url,omitempty"`
	Username   string `json:"username,omitempty"`
	Credential string `json:"credential,omitempty"`
}
type ApiResponse struct {
	Providers map[string][]ApiIceServer `json:"providers"`
	Location  string                    `json:"location"`
}

func (d *Driver) GetIceServers() (providersAndIceServers map[string][]webrtc.ICEServer, location string, err error) {
	if d.Config.RequestUrl != "" {

		client := &http.Client{}

		req, err := http.NewRequest("POST", d.Config.RequestUrl, nil)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+d.Config.ApiKey)

		if err != nil {
			// log.WithFields(log.Fields{
			// 	"error": err,
			// }).Error("Error forming http request")
			return nil, "", err
		}

		res, err := client.Do(req)
		if err != nil {
			// log.WithFields(log.Fields{
			// 	"error": err,
			// }).Error("Error doing http response")
			return nil, "", err
		}

		defer res.Body.Close()
		//check the code of the response
		if res.StatusCode != 201 {
			err = errors.New("error from our api")
			// log.WithFields(log.Fields{
			// 	"code": res.StatusCode,
			// 	"error": err,
			// }).Error("Error status code http response")
			return nil, "", err
		}

		responseData, err := io.ReadAll(res.Body)
		if err != nil {
			// log.WithFields(log.Fields{
			// 	"error": err,
			// }).Error("Error reading http response")
			return nil, "", err
		}
		// log.Info("got a response back from cloudflare api")

		responseServers := ApiResponse{}
		json.Unmarshal([]byte(responseData), &responseServers)

		// log.WithFields(log.Fields{
		// 	"response": responseServers,
		// }).Info("http response")

		location = responseServers.Location

		for k, q := range responseServers.Providers {
			iceServers := []webrtc.ICEServer{}
			for _, r := range q {

				s := webrtc.ICEServer{
					URLs: []string{r.URL},
				}

				if r.Username != "" {
					s.Username = r.Username
				}
				if r.Credential != "" {
					s.Credential = r.Credential
				}
				iceServers = append(iceServers, s)
			}
			providersAndIceServers[k] = iceServers
		}
	}
	return
}
