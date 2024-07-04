package twilio

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/nimbleape/iceperf-agent/config"
	"github.com/pion/stun/v2"
	"github.com/pion/webrtc/v4"
	// log "github.com/sirupsen/logrus"
)

type Driver struct {
	Config *config.ICEConfig
}

type TwilioIceServers struct {
	URL        string `json:"url,omitempty"`
	URLs       string `json:"urls,omitempty"`
	Username   string `json:"username,omitempty"`
	Credential string `json:"credential,omitempty"`
}
type TwilioResponse struct {
	Username    string             `json:"username"`
	DateUpdated string             `json:"date_updated"`
	TTL         string             `json:"ttl"`
	DateCreated string             `json:"date_created"`
	Password    string             `json:"password"`
	IceServers  []TwilioIceServers `json:"ice_servers"`
}

func (d *Driver) GetIceServers() (iceServers []webrtc.ICEServer, err error) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", d.Config.RequestUrl, nil)
	req.SetBasicAuth(d.Config.HttpUsername, d.Config.HttpPassword)

	if err != nil {
		// log.WithFields(log.Fields{
		// 	"error": err,
		// }).Error("Error forming http request")
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		// log.WithFields(log.Fields{
		// 	"error": err,
		// }).Error("Error doing http response")
		return nil, err
	}

	defer res.Body.Close()
	//check the code of the response
	if res.StatusCode != 201 {
		err = errors.New("error from twilio api")
		// log.WithFields(log.Fields{
		// 	"error": err,
		// }).Error("Error status code http response")
		return nil, err
	}

	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		// log.WithFields(log.Fields{
		// 	"error": err,
		// }).Error("Error reading http response")
		return nil, err
	}

	responseServers := TwilioResponse{}
	json.Unmarshal([]byte(responseData), &responseServers)

	tempTurnHost := ""

	for _, r := range responseServers.IceServers {

		info, err := stun.ParseURI(r.URL)

		if err != nil {
			return nil, err
		}

		if ((info.Scheme == stun.SchemeTypeTURN || info.Scheme == stun.SchemeTypeTURNS) && !d.Config.TurnEnabled) || ((info.Scheme == stun.SchemeTypeSTUN || info.Scheme == stun.SchemeTypeSTUNS) && !d.Config.StunEnabled) {
			continue
		}

		if info.Scheme == stun.SchemeTypeTURN {
			tempTurnHost = info.Host
		}

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

	//apparently if you go and make a tls turn uri it will work
	s := webrtc.ICEServer{
		URLs: []string{"turns:" + tempTurnHost + ":5349?transport=tcp"},
	}

	if responseServers.Username != "" {
		s.Username = responseServers.Username
	}
	if responseServers.Password != "" {
		s.Credential = responseServers.Password
	}

	iceServers = append(iceServers, s)

	return iceServers, nil
}
