package client

import (
	"fmt"
	"log/slog"

	"github.com/nimbleape/iceperf-agent/adapters"
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
	"github.com/rs/xid"
	// log "github.com/sirupsen/logrus"
)

func formGenericIceServers(config *config.ICEConfig) (adapters.IceServersConfig, error) {
	iceServers := []webrtc.ICEServer{}
	if config.StunEnabled {
		for proto, ports := range config.StunPorts {
			query := ""
			if !config.StunUseRFC7094URI {
				query = fmt.Sprintf("?transport=%s", proto)
			}
			for _, port := range ports {
				stunProto := "stun"
				if proto == "tls" {
					stunProto = "stuns"
				}
				url := fmt.Sprintf("%s:%s:%d%s", stunProto, config.StunHost, port, query)

				iceServers = append(iceServers,
					webrtc.ICEServer{
						URLs:           []string{url},
						Username:       config.Username,
						Credential:     config.Password,
						CredentialType: webrtc.ICECredentialTypePassword,
					})

			}
		}
	}
	if config.TurnEnabled {
		for proto, ports := range config.TurnPorts {
			for _, port := range ports {
				turnProto := "turn"
				l4proto := proto
				if proto == "tls" {
					turnProto = "turns"
					l4proto = "tcp"
				}
				if proto == "dtls" {
					turnProto = "turns"
					l4proto = "udp"
				}
				url := fmt.Sprintf("%s:%s:%d?transport=%s",
					turnProto, config.TurnHost, port, l4proto)

				iceServers = append(iceServers,
					webrtc.ICEServer{
						URLs:           []string{url},
						Username:       config.Username,
						Credential:     config.Password,
						CredentialType: webrtc.ICECredentialTypePassword,
					})
			}
		}

	}
	c := adapters.IceServersConfig{
		IceServers:   iceServers,
		DoThroughput: config.DoThroughput,
	}

	return c, nil
}

type IceServersConfig struct {
	IceServers   map[string][]webrtc.ICEServer
	DoThroughput bool
}

func GetIceServers(config *config.Config, logger *slog.Logger, testRunId xid.ID) (map[string]adapters.IceServersConfig, string, error) {

	//check if the API is set and is enabled
	if apiConfig, ok := config.ICEConfig["api"]; ok && apiConfig.Enabled {
		md := api.Driver{
			Config: &apiConfig,
			Logger: logger,
		}
		iceServers, node, err := md.GetIceServers(testRunId)
		return iceServers, node, err
	}

	iceServers := make(map[string]adapters.IceServersConfig)

	//loop through
	for key, conf := range config.ICEConfig {
		switch key {
		case "api":
			continue
		case "elixir":
			if !conf.Enabled {
				continue
			}
			md := elixir.Driver{
				Config: &conf,
				Logger: logger,
			}
			is, err := md.GetIceServers()
			if err != nil {
				logger.Error("Error getting elixir ice servers")
				return nil, "", err
			}
			logger.Info("elixir IceServers", "is", is)
			iceServers[key] = is
		case "google":
			if !conf.Enabled {
				continue
			}
			md := google.Driver{
				Config: &conf,
				Logger: logger,
			}
			is, err := md.GetIceServers()
			if err != nil {
				logger.Error("Error getting google ice servers")
				return nil, "", err
			}
			logger.Info("google IceServers", "is", is)

			iceServers[key] = is
		case "metered":
			if !conf.Enabled {
				continue
			}
			md := metered.Driver{
				Config: &conf,
				Logger: logger,
			}
			is, err := md.GetIceServers()
			if err != nil {
				logger.Error("Error getting metered ice servers")
				return nil, "", err
			}
			logger.Info("metered IceServers", "is", is)

			iceServers[key] = is
		case "twilio":
			if !conf.Enabled {
				continue
			}
			td := twilio.Driver{
				Config: &conf,
				Logger: logger,
			}
			is, err := td.GetIceServers()
			if err != nil {
				logger.Error("Error getting twilio ice servers")
				return nil, "", err
			}
			logger.Info("twilio IceServers", "is", is)

			iceServers[key] = is
		case "xirsys":
			if !conf.Enabled {
				continue
			}
			xd := xirsys.Driver{
				Config: &conf,
				Logger: logger,
			}
			is, err := xd.GetIceServers()
			if err != nil {
				logger.Error("Error getting xirsys ice servers")
				return nil, "", err
			}
			logger.Info("xirsys IceServers", "is", is)

			iceServers[key] = is
		case "cloudflare":
			if !conf.Enabled {
				continue
			}
			cd := cloudflare.Driver{
				Config: &conf,
				Logger: logger,
			}
			is, err := cd.GetIceServers()
			if err != nil {
				logger.Error("Error getting cloudflare ice servers")
				return nil, "", err
			}
			logger.Info("cloudflare IceServers", "is", is)

			iceServers[key] = is
		case "expressturn":
			if !conf.Enabled {
				continue
			}
			ed := expressturn.Driver{
				Config: &conf,
				Logger: logger,
			}
			is, err := ed.GetIceServers()
			if err != nil {
				logger.Error("Error getting expressturn ice servers")

				return nil, "", err
			}
			logger.Info("expressturn IceServers", "is", is)

			iceServers[key] = is
		default:
			is, err := formGenericIceServers(&conf)
			if err != nil {
				logger.Error("Error getting generic ice servers")
				return nil, "", err
			}
			logger.Info("default IceServers", "key", key, "is", is)

			iceServers[key] = is
		}
	}

	return iceServers, "", nil
}
