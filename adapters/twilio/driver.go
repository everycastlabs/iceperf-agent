package twilio

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
	// log "github.com/sirupsen/logrus"
)

type Driver struct {
	Config *config.ICEConfig
	Logger *slog.Logger
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

func (d *Driver) GetIceServers() (adapters.IceServersConfig, error) {
	client := &http.Client{}

	iceServers := adapters.IceServersConfig{
		IceServers: []webrtc.ICEServer{},
	}

	req, err := http.NewRequest("POST", d.Config.RequestUrl, nil)
	req.SetBasicAuth(d.Config.HttpUsername, d.Config.HttpPassword)

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
	if res.StatusCode != 201 {
		err = errors.New("error from twilio api")
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

	responseServers := TwilioResponse{}
	json.Unmarshal([]byte(responseData), &responseServers)

	tempTurnHost := ""

	gotTransports := make(map[string]bool)

	gotTransports[stun.SchemeTypeSTUN.String()+stun.ProtoTypeUDP.String()] = false
	gotTransports[stun.SchemeTypeSTUN.String()+stun.ProtoTypeTCP.String()] = false
	gotTransports[stun.SchemeTypeTURN.String()+stun.ProtoTypeUDP.String()] = false
	gotTransports[stun.SchemeTypeTURN.String()+stun.ProtoTypeTCP.String()] = false
	gotTransports[stun.SchemeTypeTURNS.String()+stun.ProtoTypeTCP.String()] = false

	for _, r := range responseServers.IceServers {

		info, err := stun.ParseURI(r.URL)

		if err != nil {
			return iceServers, err
		}

		if ((info.Scheme == stun.SchemeTypeTURN || info.Scheme == stun.SchemeTypeTURNS) && !d.Config.TurnEnabled) || ((info.Scheme == stun.SchemeTypeSTUN || info.Scheme == stun.SchemeTypeSTUNS) && !d.Config.StunEnabled) {
			continue
		}

		if gotTransports[info.Scheme.String()+info.Proto.String()] {
			//we don't want to test all the special ports right now
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
		iceServers.IceServers = append(iceServers.IceServers, s)
		gotTransports[info.Scheme.String()+info.Proto.String()] = true
	}

	if d.Config.TurnEnabled {
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

		iceServers.IceServers = append(iceServers.IceServers, s)
	}

	return iceServers, nil
}
