package xirsys

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

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

func (d *Driver) GetIceServers() (adapters.IceServersConfig, error) {
	client := &http.Client{}

	iceServers := adapters.IceServersConfig{
		IceServers: []webrtc.ICEServer{},
	}

	req, err := http.NewRequest("PUT", d.Config.RequestUrl, strings.NewReader(`{"format": "urls", "expire": "1800"}`))
	req.SetBasicAuth(d.Config.HttpUsername, d.Config.HttpPassword)
	req.Header.Add("Content-Type", "application/json")

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
		err = errors.New("error from xirsys api")
		// log.WithFields(log.Fields{
		// 	"error":  err,
		// 	"status": res.StatusCode,
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

	responseServers := XirsysResponse{}
	json.Unmarshal([]byte(responseData), &responseServers)

	gotTransports := make(map[string]bool)

	gotTransports[stun.SchemeTypeSTUN.String()+stun.ProtoTypeUDP.String()] = false
	gotTransports[stun.SchemeTypeSTUN.String()+stun.ProtoTypeTCP.String()] = false
	gotTransports[stun.SchemeTypeTURN.String()+stun.ProtoTypeUDP.String()] = false
	gotTransports[stun.SchemeTypeTURN.String()+stun.ProtoTypeTCP.String()] = false
	gotTransports[stun.SchemeTypeTURNS.String()+stun.ProtoTypeTCP.String()] = false

	for _, r := range responseServers.V.IceServers.URLs {

		info, err := stun.ParseURI(r)

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

		s := webrtc.ICEServer{
			URLs: []string{r},
		}

		if responseServers.V.IceServers.Username != "" {
			s.Username = responseServers.V.IceServers.Username
		}
		if responseServers.V.IceServers.Credential != "" {
			s.Credential = responseServers.V.IceServers.Credential
		}
		iceServers.IceServers = append(iceServers.IceServers, s)
		gotTransports[info.Scheme.String()+info.Proto.String()] = true
	}

	return iceServers, nil
}
