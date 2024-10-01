package adapters

import "github.com/pion/webrtc/v4"

type IceServersConfig struct {
	IceServers   []webrtc.ICEServer
	DoThroughput bool
	TurnToTurn   bool
}
