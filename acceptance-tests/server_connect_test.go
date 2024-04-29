package acceptance_tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/joho/godotenv"
	"github.com/nimbleape/iceperf-agent/adapters/serverconnect"
	"github.com/nimbleape/iceperf-agent/adapters/webrtcpeerconnect"
	"github.com/nimbleape/iceperf-agent/specifications"
	"github.com/pion/webrtc/v4"
)

// TODO remove
func loadEnv() error {
	ENVIRONMENT := os.Getenv("ENVIRONMENT")
	if ENVIRONMENT != "TEST" {
		err := godotenv.Load("../.env")
		return err
	}
	return nil
}

// FIXME this is actually just a "Connect to TURN provider" test
func TestConnectToServer(t *testing.T) {
	// FIXME use new config
	err := loadEnv()
	assert.NoError(t, err)

	API_KEY := os.Getenv("METERED_API_KEY")
	driver := serverconnect.Driver{
		Url: fmt.Sprintf("https://relayperf.metered.live/api/v1/turn/credentials?apiKey=%s", API_KEY),
		Client: &http.Client{
			Timeout: 1 * time.Second,
		},
	}

	specifications.ConnectToServerSpecification(t, driver)
}

func TestConnectToTURNServer(t *testing.T) {
	// FIXME use new config
	err := loadEnv()
	assert.NoError(t, err)

	USERNAME := os.Getenv("METERED_USERNAME")
	PASSWORD := os.Getenv("METERED_PASSWORD")
	driver := webrtcpeerconnect.Driver{
		ICEServers: []webrtc.ICEServer{
			{
				URLs:       []string{"turn:standard.relay.metered.ca:80"},
				Username:   USERNAME,
				Credential: PASSWORD,
			},
		}, // TODO more servers
		ICETransportPolicy: webrtc.ICETransportPolicyRelay,
	}

	specifications.ConnectToServerSpecification(t, driver)
}

// TODO perhaps test connecting round trip offerer and answerer
