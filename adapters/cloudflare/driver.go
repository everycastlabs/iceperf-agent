package cloudflare

import (
	"fmt"

	"github.com/nimbleape/iceperf-agent/config"
	"github.com/pion/webrtc/v4"
)

type Driver struct {
	Config *config.ICEConfig
}

func (d *Driver) GetIceServers() (iceServers []webrtc.ICEServer, err error) {
	if d.Config.StunHost != "" && d.Config.StunEnabled {
		if _, ok := d.Config.StunPorts["udp"]; ok {
			for _, port := range d.Config.StunPorts["udp"] {
				iceServers = append(iceServers, webrtc.ICEServer{
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
					iceServers = append(iceServers, webrtc.ICEServer{
						URLs:       []string{fmt.Sprintf("turn:%s:%d?transport=udp", d.Config.TurnHost, port)},
						Username:   d.Config.Username,
						Credential: d.Config.Password,
					})
				}
			case "tcp":
				for _, port := range d.Config.TurnPorts["tcp"] {
					iceServers = append(iceServers, webrtc.ICEServer{
						URLs:       []string{fmt.Sprintf("turn:%s:%d?transport=tcp", d.Config.TurnHost, port)},
						Username:   d.Config.Username,
						Credential: d.Config.Password,
					})
				}
			case "tls":
				for _, port := range d.Config.TurnPorts["tls"] {
					iceServers = append(iceServers, webrtc.ICEServer{
						URLs:       []string{fmt.Sprintf("turns:%s:%d?transport=tcp", d.Config.TurnHost, port)},
						Username:   d.Config.Username,
						Credential: d.Config.Password,
					})
				}
			default:
			}
		}
	}
	return
}
