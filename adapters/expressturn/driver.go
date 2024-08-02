package expressturn

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
	return iceServers, nil
}
