package google

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
	return
}
