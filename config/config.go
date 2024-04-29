package config

import (
	"github.com/pion/webrtc/v4"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type ICEConfig struct {
	Username     string           `yaml:"username,omitempty"`
	Password     string           `yaml:"password,omitempty"`
	ApiKey       string           `yaml:"api_key,omitempty"`
	AccountSid   string           `yaml:"account_sid,omitempty"`
	RequestUrl   string           `yaml:"request_url,omitempty"`
	HttpUsername string           `yaml:"http_username"`
	HttpPassword string           `yaml:"http_password"`
	Enabled      bool             `yaml:"enabled"`
	StunHost     string           `yaml:"stun_host,omitempty"`
	TurnHost     string           `yaml:"turn_host,omitempty"`
	TurnPorts    map[string][]int `yaml:"turn_ports,omitempty"`
	StunPorts    map[string][]int `yaml:"stun_ports,omitempty"`
	StunEnabled  bool             `yaml:"stun_enabled"`
	TurnEnabled  bool             `yaml:"turn_enabled"`
}

type LokiConfig struct {
	Enabled        bool              `yaml:"enabled"`
	UseBasicAuth   bool              `yaml:"use_basic_auth"`
	UseHeadersAuth bool              `yaml:"use_headers_auth"`
	Username       string            `yaml:"username,omitempty"`
	Password       string            `yaml:"password,omitempty"`
	URL            string            `yaml:"url"`
	AuthHeaders    map[string]string `yaml:"auth_headers,omitempty"`
}

type LoggingConfig struct {
	Level string     `yaml:"level"`
	Loki  LokiConfig `yaml:"loki"`
}

type Config struct {
	ICEConfig map[string]ICEConfig `yaml:"ice_servers"`
	Logging   LoggingConfig        `yaml:"logging"`

	WebRTCConfig webrtc.Configuration
	// TODO the following should be different for answerer and offerer sides
	OnICECandidate          func(*webrtc.ICECandidate)
	OnConnectionStateChange func(s webrtc.PeerConnectionState)

	// internal
	ServiceName string `yaml:"-"`
	NodeID      string // Do not provide, will be overwritten
	Logger      *logrus.Entry
}

func NewConfig(confString string) (*Config, error) {
	c := &Config{
		ServiceName: "ICEPerf",
	}
	if confString != "" {
		if err := yaml.Unmarshal([]byte(confString), c); err != nil {
			return nil, err
		}
	}
	return c, nil
}
