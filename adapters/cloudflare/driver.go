package cloudflare

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/nimbleape/iceperf-agent/adapters"
	"github.com/nimbleape/iceperf-agent/config"
	"github.com/pion/stun/v2"
	"github.com/pion/webrtc/v4"
)

type Driver struct {
	Config *config.ICEConfig
	Logger *slog.Logger
}

type CloudflareIceServers struct {
	URLs       []string `json:"urls,omitempty"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

type CloudflareResponse struct {
	IceServers CloudflareIceServers `json:"iceServers"`
}

func (d *Driver) GetIceServers() (adapters.IceServersConfig, error) {

	iceServers := adapters.IceServersConfig{
		IceServers:   []webrtc.ICEServer{},
		DoThroughput: d.Config.DoThroughput,
	}

	if d.Config.RequestUrl != "" {

		client := &http.Client{}

		req, err := http.NewRequest("POST", d.Config.RequestUrl, strings.NewReader(`{"ttl": 86400}`))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+d.Config.ApiKey)

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
			err = errors.New("error from cloudflare api")
			// log.WithFields(log.Fields{
			// 	"code": res.StatusCode,
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
		// log.Info("got a response back from cloudflare api")

		responseServers := CloudflareResponse{}
		json.Unmarshal([]byte(responseData), &responseServers)

		// log.WithFields(log.Fields{
		// 	"response": responseServers,
		// }).Info("http response")

		for _, r := range responseServers.IceServers.URLs {

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

			if responseServers.IceServers.Username != "" {
				s.Username = responseServers.IceServers.Username
			}
			if responseServers.IceServers.Credential != "" {
				s.Credential = responseServers.IceServers.Credential
			}
			iceServers.IceServers = append(iceServers.IceServers, s)
		}
	} else {
		if d.Config.StunHost != "" && d.Config.StunEnabled {
			if _, ok := d.Config.StunPorts["udp"]; ok {
				for _, port := range d.Config.StunPorts["udp"] {
					iceServers.IceServers = append(iceServers.IceServers, webrtc.ICEServer{
						URLs: []string{fmt.Sprintf("stun:%s:%d", d.Config.StunHost, port)},
					})
				}
			}
		}

		if d.Config.TurnHost != "" && d.Config.TurnEnabled {
			for transport := range d.Config.TurnPorts {
				switch transport {
				case "udp":
					for _, port := range d.Config.TurnPorts["udp"] {
						iceServers.IceServers = append(iceServers.IceServers, webrtc.ICEServer{
							URLs:       []string{fmt.Sprintf("turn:%s:%d?transport=udp", d.Config.TurnHost, port)},
							Username:   d.Config.Username,
							Credential: d.Config.Password,
						})
					}
				case "tcp":
					for _, port := range d.Config.TurnPorts["tcp"] {
						iceServers.IceServers = append(iceServers.IceServers, webrtc.ICEServer{
							URLs:       []string{fmt.Sprintf("turn:%s:%d?transport=tcp", d.Config.TurnHost, port)},
							Username:   d.Config.Username,
							Credential: d.Config.Password,
						})
					}
				case "tls":
					for _, port := range d.Config.TurnPorts["tls"] {
						iceServers.IceServers = append(iceServers.IceServers, webrtc.ICEServer{
							URLs:       []string{fmt.Sprintf("turns:%s:%d?transport=tcp", d.Config.TurnHost, port)},
							Username:   d.Config.Username,
							Credential: d.Config.Password,
						})
					}
				default:
				}
			}
		}
	}
	return iceServers, nil
}
