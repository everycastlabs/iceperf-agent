package elixir

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/nimbleape/iceperf-agent/adapters"
	"github.com/nimbleape/iceperf-agent/config"
	"github.com/pion/stun/v2"
	"github.com/pion/webrtc/v4"
)

type Driver struct {
	Config *config.ICEConfig
	Logger *slog.Logger
}

type ElixirResponse struct {
	Username   string   `json:"username"`
	TTL        string   `json:"ttl"`
	Password   string   `json:"password"`
	IceServers []string `json:"uris"`
}

// func (d Driver) Measure(measurementName string) error {
// 	return nil
// }

func (d *Driver) GetIceServers() (adapters.IceServersConfig, error) {

	iceServers := adapters.IceServersConfig{
		IceServers:   []webrtc.ICEServer{},
		DoThroughput: d.Config.DoThroughput,
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", d.Config.RequestUrl+"&username="+d.Config.HttpUsername, nil)

	if err != nil {
		// log.WithFields(log.Fields{
		// 	"error": err,
		// }).Error("Error forming http request")
		return iceServers, err
	}

	res, err := client.Do(req)
	if err != nil {
		// log.WithFields(log.Fields{
		// 	"error": err,
		// }).Error("Error doing http response")
		return iceServers, err
	}

	defer res.Body.Close()
	//check the code of the response
	if res.StatusCode != 200 {
		err = errors.New("error from elixir api")
		// log.WithFields(log.Fields{
		// 	"error": err,
		// }).Error("Error status code http response")
		return iceServers, err
	}

	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		// log.WithFields(log.Fields{
		// 	"error": err,
		// }).Error("Error reading http response")
		return iceServers, err
	}

	responseServers := ElixirResponse{}
	json.Unmarshal([]byte(responseData), &responseServers)

	for _, r := range responseServers.IceServers {

		info, err := stun.ParseURI(r)

		if err != nil {
			return iceServers, err
		}

		if ((info.Scheme == stun.SchemeTypeTURN || info.Scheme == stun.SchemeTypeTURNS) && !d.Config.TurnEnabled) || ((info.Scheme == stun.SchemeTypeSTUN || info.Scheme == stun.SchemeTypeSTUNS) && !d.Config.StunEnabled) {
			continue
		}

		s := webrtc.ICEServer{
			URLs: []string{r},
		}

		if responseServers.Username != "" {
			s.Username = responseServers.Username
		}
		if responseServers.Password != "" {
			s.Credential = responseServers.Password
		}

		iceServers.IceServers = append(iceServers.IceServers, s)

		if d.Config.StunEnabled {
			stun := webrtc.ICEServer{
				URLs: []string{"stun:" + info.Host + ":3478"},
			}
			iceServers.IceServers = append(iceServers.IceServers, stun)
		}
	}

	//

	return iceServers, nil
}
