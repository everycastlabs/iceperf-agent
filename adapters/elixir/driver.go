package elixir

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/nimbleape/iceperf-agent/config"
	"github.com/pion/stun/v2"
	"github.com/pion/webrtc/v4"
	log "github.com/sirupsen/logrus"
)

type Driver struct {
	Config *config.ICEConfig
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

func (d *Driver) GetIceServers() (iceServers []webrtc.ICEServer, err error) {
	client := &http.Client{}

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error forming http request URL")
		return nil, err
	}
	req, err := http.NewRequest("POST", d.Config.RequestUrl+"&username="+d.Config.HttpUsername, nil)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error forming http request")
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error doing http response")
		return nil, err
	}

	defer res.Body.Close()
	//check the code of the response
	if res.StatusCode != 200 {
		err = errors.New("error from elixir api")
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error status code http response")
		return nil, err
	}

	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error reading http response")
		return nil, err
	}

	responseServers := ElixirResponse{}
	json.Unmarshal([]byte(responseData), &responseServers)

	for _, r := range responseServers.IceServers {

		info, err := stun.ParseURI(r)

		if err != nil {
			return nil, err
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

		iceServers = append(iceServers, s)

		if d.Config.StunEnabled {
			stun := webrtc.ICEServer{
				URLs: []string{"stun:" + info.Host + ":3478"},
			}
			iceServers = append(iceServers, stun)
		}
	}

	//

	return iceServers, nil
}
