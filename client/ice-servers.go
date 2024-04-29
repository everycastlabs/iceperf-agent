package client

import (
	"github.com/nimbleape/go-relay-perf-com-tests/adapters/cloudflare"
	"github.com/nimbleape/go-relay-perf-com-tests/adapters/elixir"
	"github.com/nimbleape/go-relay-perf-com-tests/adapters/expressturn"
	"github.com/nimbleape/go-relay-perf-com-tests/adapters/google"
	"github.com/nimbleape/go-relay-perf-com-tests/adapters/metered"
	"github.com/nimbleape/go-relay-perf-com-tests/adapters/twilio"
	"github.com/nimbleape/go-relay-perf-com-tests/adapters/xirsys"
	"github.com/nimbleape/go-relay-perf-com-tests/config"
	"github.com/pion/webrtc/v4"
	log "github.com/sirupsen/logrus"
)

func formGenericIceServers(config *config.ICEConfig) (iceServers []webrtc.ICEServer, err error) {
	return nil, nil
}

func GetIceServers(config *config.Config) (iceServers map[string][]webrtc.ICEServer, err error) {
	iceServers = make(map[string][]webrtc.ICEServer)

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
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Error getting elixir ice servers")
				return nil, err
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
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Error getting google ice servers")
				return nil, err
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
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Error getting metered ice servers")
				return nil, err
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
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Error getting twilio ice servers")
				return nil, err
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
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Error getting xirsys ice servers")
				return nil, err
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
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Error getting Cloudflare ice servers")
				return nil, err
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
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Error getting Expressturn ice servers")
				return nil, err
			}
			iceServers[key] = is
		default:
			is, err := formGenericIceServers(&conf)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Error constructing ice servers")
				return nil, err
			}
			iceServers[key] = is
		}
	}

	return
}
