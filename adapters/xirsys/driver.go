package xirsys

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/nimbleape/go-relay-perf-com-tests/config"
	"github.com/pion/stun/v2"
	"github.com/pion/webrtc/v4"
	log "github.com/sirupsen/logrus"
)

type Driver struct {
	Config *config.ICEConfig
}

type XirsysIceServers struct {
	URLs       []string `json:"urls,omitempty"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

type XirsysIceServer struct {
	IceServers XirsysIceServers `json:"iceServers"`
}
type XirsysResponse struct {
	V XirsysIceServer `json:"v"`
	S string          `json:"s"`
}

func (d *Driver) GetIceServers() (iceServers []webrtc.ICEServer, err error) {
	client := &http.Client{}

	req, err := http.NewRequest("PUT", d.Config.RequestUrl, strings.NewReader(`{"format": "urls"}`))
	req.SetBasicAuth(d.Config.HttpUsername, d.Config.HttpPassword)
	req.Header.Add("Content-Type", "application/json")

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
		err = errors.New("error from xirsys api")
		log.WithFields(log.Fields{
			"error":  err,
			"status": res.StatusCode,
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

	responseServers := XirsysResponse{}
	json.Unmarshal([]byte(responseData), &responseServers)

	for _, r := range responseServers.V.IceServers.URLs {

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

		if responseServers.V.IceServers.Username != "" {
			s.Username = responseServers.V.IceServers.Username
		}
		if responseServers.V.IceServers.Credential != "" {
			s.Credential = responseServers.V.IceServers.Credential
		}
		iceServers = append(iceServers, s)
	}

	return iceServers, nil
}
