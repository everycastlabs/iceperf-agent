package specifications

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/nimbleape/iceperf-agent/adapters"
)

/*
The first specification would be tht the user can connect
to a TURN server (initially just one, more options in the future).
The AT should call the server and confirm that we have connected.

The talking to the server will be handled by a Driver.
*/

type ServerConnect interface {
	Connect() (bool, error)
}
type TURNProvider interface {
	GetIceServers() (adapters.IceServersConfig, error)
}

func ConnectToServerSpecification(t testing.TB, serverConnect ServerConnect) {
	connected, err := serverConnect.Connect()
	assert.NoError(t, err)
	assert.True(t, connected)
}

func GetIceServersSpecification(t testing.TB, provider TURNProvider) {
	is, err := provider.GetIceServers()
	assert.NoError(t, err)
	assert.True(t, len(is.IceServers) > 0)
}
