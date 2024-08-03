package config

import (
	"log/slog"

	"github.com/pion/webrtc/v4"
	"github.com/prometheus/client_golang/prometheus"
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
	DoThroughput bool             `yaml:"do_throughput"`
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

type PromConfig struct {
	Enabled     bool              `yaml:"enabled"`
	URL         string            `yaml:"url"`
	AuthHeaders map[string]string `yaml:"auth_headers,omitempty"`
}

type ApiConfig struct {
	Enabled bool   `yaml:"enabled"`
	URI     string `yaml:"uri"`
	ApiKey  string `yaml:"api_key,omitempty"`
}

type LoggingConfig struct {
	Level      string     `yaml:"level"`
	API        ApiConfig  `yaml:"api"`
	Loki       LokiConfig `yaml:"loki"`
	Prometheus PromConfig `yaml:"prometheus"`
}

type TimerConfig struct {
	Enabled  bool `yaml:"enabled"`
	Interval int  `yaml:"interval"`
}

type Config struct {
	NodeID    string               `yaml:"node_id"`
	ICEConfig map[string]ICEConfig `yaml:"ice_servers"`
	Logging   LoggingConfig        `yaml:"logging"`
	Timer     TimerConfig          `yaml:"timer"`

	WebRTCConfig webrtc.Configuration
	// TODO the following should be different for answerer and offerer sides
	OnICECandidate          func(*webrtc.ICECandidate)
	OnConnectionStateChange func(s webrtc.PeerConnectionState)

	// internal
	ServiceName string `yaml:"-"`
	Logger      *slog.Logger
	Registry    *prometheus.Registry
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
