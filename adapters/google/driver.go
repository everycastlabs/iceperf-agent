package google

import (
	"fmt"
	"log/slog"

	"github.com/nimbleape/iceperf-agent/adapters"
	"github.com/nimbleape/iceperf-agent/config"
	"github.com/pion/webrtc/v4"
)

type Driver struct {
	Config *config.ICEConfig
	Logger *slog.Logger
}

func (d *Driver) GetIceServers() (adapters.IceServersConfig, error) {

	iceServers := adapters.IceServersConfig{
		IceServers: []webrtc.ICEServer{},
	}

	if d.Config.StunHost != "" && d.Config.StunEnabled {
		if _, ok := d.Config.StunPorts["udp"]; ok {
			for _, port := range d.Config.StunPorts["udp"] {
				iceServers.IceServers = append(iceServers.IceServers, webrtc.ICEServer{
					URLs: []string{fmt.Sprintf("stun:%s:%d", d.Config.StunHost, port)},
				})
			}
		}
	}
	return iceServers, nil
}
