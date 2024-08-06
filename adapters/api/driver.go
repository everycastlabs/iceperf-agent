package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/nimbleape/iceperf-agent/adapters"
	"github.com/nimbleape/iceperf-agent/config"
	"github.com/pion/webrtc/v4"
	"github.com/rs/xid"
)

type Driver struct {
	Config *config.ICEConfig
	Logger *slog.Logger
}

type ApiIceServer struct {
	URLs       []string `json:"urls,omitempty"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

type ProviderRes struct {
	IceServers   []ApiIceServer `json:"iceServers"`
	DoThroughput bool           `json:"doThroughput"`
}
type ApiResponse struct {
	Providers map[string]ProviderRes `json:"providers"`
	Node      string                 `json:"node"`
}

func (d *Driver) GetIceServers(testRunId xid.ID) (map[string]adapters.IceServersConfig, string, error) {
	providersAndIceServers := make(map[string]adapters.IceServersConfig)

	if d.Config.RequestUrl != "" {

		client := &http.Client{}

		req, err := http.NewRequest("POST", d.Config.RequestUrl, strings.NewReader(`{"testRunID": "`+testRunId.String()+`"}`))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+d.Config.ApiKey)

		if err != nil {
			// log.WithFields(log.Fields{
			// 	"error": err,
			// }).Error("Error forming http request")
			return providersAndIceServers, "", err
		}

		res, err := client.Do(req)
		if err != nil {
			// log.WithFields(log.Fields{
			// 	"error": err,
			// }).Error("Error doing http response")
			return providersAndIceServers, "", err
		}

		defer res.Body.Close()
		//check the code of the response
		if res.StatusCode != 200 {
			err = errors.New("error from our api")
			return providersAndIceServers, "", err
		}

		responseData, err := io.ReadAll(res.Body)
		if err != nil {
			// log.WithFields(log.Fields{
			// 	"error": err,
			// }).Error("Error reading http response")
			return providersAndIceServers, "", err
		}
		// log.Info("got a response back from cloudflare api")

		responseServers := ApiResponse{}
		json.Unmarshal([]byte(responseData), &responseServers)

		// log.WithFields(log.Fields{
		// 	"response": responseServers,
		// }).Info("http response")

		node := responseServers.Node

		for k, q := range responseServers.Providers {

			iceServers := []webrtc.ICEServer{}
			for _, r := range q.IceServers {

				s := webrtc.ICEServer{
					URLs: r.URLs,
				}

				if r.Username != "" {
					s.Username = r.Username
				}
				if r.Credential != "" {
					s.Credential = r.Credential
				}
				iceServers = append(iceServers, s)
			}
			providersAndIceServers[k] = adapters.IceServersConfig{
				DoThroughput: q.DoThroughput,
				IceServers:   iceServers,
			}
		}
		return providersAndIceServers, node, nil
	}
	return providersAndIceServers, "", nil
}
