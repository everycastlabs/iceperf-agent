package webrtcpeerconnect

import (
	"github.com/pion/webrtc/v4"
)

type Driver struct {
	ICEServers         []webrtc.ICEServer
	ICETransportPolicy webrtc.ICETransportPolicy
}

func (d Driver) Connect() (connected bool, err error) {
	// FIXME use new config
	// config := client.NewClientConfig()
	// config.WebRTCConfig.ICEServers = d.ICEServers
	// config.WebRTCConfig.ICETransportPolicy = d.ICETransportPolicy

	// c, err := client.NewClient(config)
	// if err != nil {
	// 	return false, err
	// }
	// c.Run()

	// state := <-c.OffererConnected
	// return state, nil
	return true, nil
}
