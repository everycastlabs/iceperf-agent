package client

import (
	"log/slog"

	"github.com/nimbleape/iceperf-agent/adapters/api"
	"github.com/nimbleape/iceperf-agent/adapters/cloudflare"
	"github.com/nimbleape/iceperf-agent/adapters/elixir"
	"github.com/nimbleape/iceperf-agent/adapters/expressturn"
	"github.com/nimbleape/iceperf-agent/adapters/google"
	"github.com/nimbleape/iceperf-agent/adapters/metered"
	"github.com/nimbleape/iceperf-agent/adapters/twilio"
	"github.com/nimbleape/iceperf-agent/adapters/xirsys"
	"github.com/nimbleape/iceperf-agent/config"
	"github.com/pion/webrtc/v4"
	// log "github.com/sirupsen/logrus"
)

func formGenericIceServers(config *config.ICEConfig) (iceServers []webrtc.ICEServer, err error) {
	return nil, nil
}

func GetIceServers(config *config.Config, logger *slog.Logger) (iceServers map[string][]webrtc.ICEServer, location string, err error) {
	iceServers = make(map[string][]webrtc.ICEServer)

	//check if the API is set and is enabled
	if apiConfig, ok := config.ICEConfig["api"]; ok && apiConfig.Enabled {
		md := api.Driver{
			Config: &apiConfig,
		}
		iceServers, location, err = md.GetIceServers()
		return
	}

	//loop through
	for key, conf := range config.ICEConfig {
		switch key {
		case "elixir":
			if !conf.Enabled {
				continue
			}
			md := elixir.Driver{
				Config: &conf,
			}
			is, err := md.GetIceServers()
			if err != nil {
				logger.Error("Error getting elixir ice servers")
				return nil, "", err
			}
			iceServers[key] = is
		case "google":
			if !conf.Enabled {
				continue
			}
			md := google.Driver{
				Config: &conf,
			}
			is, err := md.GetIceServers()
			if err != nil {
				logger.Error("Error getting google ice servers")
				return nil, "", err
			}
			iceServers[key] = is
		case "metered":
			if !conf.Enabled {
				continue
			}
			md := metered.Driver{
				Config: &conf,
			}
			is, err := md.GetIceServers()
			if err != nil {
				logger.Error("Error getting metered ice servers")
				return nil, "", err
			}
			iceServers[key] = is
		case "twilio":
			if !conf.Enabled {
				continue
			}
			td := twilio.Driver{
				Config: &conf,
			}
			is, err := td.GetIceServers()
			if err != nil {
				logger.Error("Error getting twilio ice servers")
				return nil, "", err
			}
			iceServers[key] = is
		case "xirsys":
			if !conf.Enabled {
				continue
			}
			xd := xirsys.Driver{
				Config: &conf,
			}
			is, err := xd.GetIceServers()
			if err != nil {
				logger.Error("Error getting xirsys ice servers")
				return nil, "", err
			}
			iceServers[key] = is
		case "cloudflare":
			if !conf.Enabled {
				continue
			}
			cd := cloudflare.Driver{
				Config: &conf,
			}
			is, err := cd.GetIceServers()
			if err != nil {
				logger.Error("Error getting cloudflare ice servers")
				return nil, "", err
			}
			iceServers[key] = is
		case "expressturn":
			if !conf.Enabled {
				continue
			}
			ed := expressturn.Driver{
				Config: &conf,
			}
			is, err := ed.GetIceServers()
			if err != nil {
				logger.Error("Error getting expressturn ice servers")

				return nil, "", err
			}
			iceServers[key] = is
		default:
			is, err := formGenericIceServers(&conf)
			if err != nil {
				logger.Error("Error getting generic ice servers")
				return nil, "", err
			}
			iceServers[key] = is
		}
	}

	return
}
