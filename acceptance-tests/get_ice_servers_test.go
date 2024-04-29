package acceptance_tests

import (
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/nimbleape/go-relay-perf-com-tests/adapters/cloudflare"
	"github.com/nimbleape/go-relay-perf-com-tests/adapters/expressturn"
	"github.com/nimbleape/go-relay-perf-com-tests/adapters/metered"
	"github.com/nimbleape/go-relay-perf-com-tests/adapters/twilio"
	"github.com/nimbleape/go-relay-perf-com-tests/adapters/xirsys"
	"github.com/nimbleape/go-relay-perf-com-tests/config"
	"github.com/nimbleape/go-relay-perf-com-tests/specifications"
)

func TestMeteredICEServers(t *testing.T) {
	err := loadEnv()
	assert.NoError(t, err)

	API_KEY := os.Getenv("METERED_API_KEY")
	REQUEST_URL := os.Getenv("METERED_REQUEST_URL")

	meteredConfig := config.ICEConfig{
		RequestUrl:  REQUEST_URL,
		ApiKey:      API_KEY,
		StunEnabled: true,
		TurnEnabled: true,
	}

	md := metered.Driver{
		Config: &meteredConfig,
	}

	specifications.GetIceServersSpecification(t, &md)
}

func TestTwilioICEServers(t *testing.T) {
	err := loadEnv()
	assert.NoError(t, err)

	HTTP_USERNAME := os.Getenv("TWILIO_HTTP_USERNAME")
	HTTP_PASSWORD := os.Getenv("TWILIO_HTTP_PASSWORD")
	REQUEST_URL := os.Getenv("TWILIO_REQUEST_URL")

	td := twilio.Driver{
		Config: &config.ICEConfig{
			RequestUrl:   REQUEST_URL,
			HttpUsername: HTTP_USERNAME,
			HttpPassword: HTTP_PASSWORD,
			StunEnabled:  true,
			TurnEnabled:  true,
		},
	}

	specifications.GetIceServersSpecification(t, &td)
}

func TestXirsysICEServers(t *testing.T) {
	err := loadEnv()
	assert.NoError(t, err)

	HTTP_USERNAME := os.Getenv("XIRSYS_HTTP_USERNAME")
	HTTP_PASSWORD := os.Getenv("XIRSYS_HTTP_PASSWORD")
	REQUEST_URL := os.Getenv("XIRSYS_REQUEST_URL")

	xd := xirsys.Driver{
		Config: &config.ICEConfig{
			RequestUrl:   REQUEST_URL,
			HttpUsername: HTTP_USERNAME,
			HttpPassword: HTTP_PASSWORD,
			StunEnabled:  true,
			TurnEnabled:  true,
		},
	}

	specifications.GetIceServersSpecification(t, &xd)
}
func TestCloudflareICEServers(t *testing.T) {
	err := loadEnv()
	assert.NoError(t, err)

	USERNAME := os.Getenv("CLOUDFLARE_USERNAME")
	PASSWORD := os.Getenv("CLOUDFLARE_PASSWORD")

	turnPorts := make(map[string][]int)
	turnPorts["udp"] = []int{3478, 53}
	turnPorts["tcp"] = []int{3478, 80}
	turnPorts["tls"] = []int{5349, 443}

	cd := cloudflare.Driver{
		Config: &config.ICEConfig{
			Username:    USERNAME,
			Password:    PASSWORD,
			StunHost:    "stun.cloudflare.com",
			TurnHost:    "turn.cloudflare.com",
			TurnPorts:   turnPorts,
			StunEnabled: true,
			TurnEnabled: true,
		},
	}

	specifications.GetIceServersSpecification(t, &cd)
}

func TestExpressturnICEServers(t *testing.T) {
	err := loadEnv()
	assert.NoError(t, err)

	USERNAME := os.Getenv("EXPRESSTURN_USERNAME")
	PASSWORD := os.Getenv("EXPRESSTURN_PASSWORD")

	turnPorts := make(map[string][]int)
	turnPorts["udp"] = []int{3478, 80}
	turnPorts["tcp"] = []int{3478, 443}
	turnPorts["tls"] = []int{5349, 443}

	ed := expressturn.Driver{
		Config: &config.ICEConfig{
			Username:    USERNAME,
			Password:    PASSWORD,
			StunHost:    "relay1.expressturn.com",
			TurnHost:    "relay1.expressturn.com",
			TurnPorts:   turnPorts,
			StunEnabled: true,
			TurnEnabled: true,
		},
	}

	specifications.GetIceServersSpecification(t, &ed)
}
