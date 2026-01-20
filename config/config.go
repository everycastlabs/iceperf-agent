package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"
)

type ICEConfig struct {
	Username          string           `yaml:"username,omitempty"`
	Password          string           `yaml:"password,omitempty"`
	ApiKey            string           `json:"apiKey,omitempty" yaml:"api_key,omitempty"`
	AccountSid        string           `yaml:"account_sid,omitempty"`
	RequestUrl        string           `json:"requestUrl,omitempty" yaml:"request_url,omitempty"`
	HttpUsername      string           `yaml:"http_username"`
	HttpPassword      string           `yaml:"http_password"`
	Enabled           bool             `yaml:"enabled"`
	StunUseRFC7094URI bool             `yaml:"stun_use_rfc7094_uri"`
	StunHost          string           `yaml:"stun_host,omitempty"`
	TurnHost          string           `yaml:"turn_host,omitempty"`
	TurnPorts         map[string][]int `yaml:"turn_ports,omitempty"`
	StunPorts         map[string][]int `yaml:"stun_ports,omitempty"`
	StunEnabled       bool             `yaml:"stun_enabled"`
	TurnEnabled       bool             `yaml:"turn_enabled"`
	DoThroughput      bool             `yaml:"do_throughput"`
}

type LokiConfig struct {
	Enabled        bool              `json:"enabled" yaml:"enabled"`
	UseBasicAuth   bool              `yaml:"use_basic_auth"`
	UseHeadersAuth bool              `yaml:"use_headers_auth"`
	Username       string            `yaml:"username,omitempty"`
	Password       string            `yaml:"password,omitempty"`
	URL            string            `json:"url" yaml:"url"`
	AuthHeaders    map[string]string `yaml:"auth_headers,omitempty"`
}

type PromConfig struct {
	Enabled     bool              `yaml:"enabled"`
	URL         string            `yaml:"url"`
	AuthHeaders map[string]string `yaml:"auth_headers,omitempty"`
}

type ApiConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	URI     string `json:"uri" yaml:"uri"`
	ApiKey  string `json:"apiKey,omitempty" yaml:"api_key,omitempty"`
}

type LoggingConfig struct {
	Level      string     `yaml:"level"`
	API        ApiConfig  `json:"api" yaml:"api"`
	Loki       LokiConfig `json:"loki" yaml:"loki"`
	Prometheus PromConfig `yaml:"prometheus"`
}

type TimerConfig struct {
	Enabled  bool `json:"enabled" yaml:"enabled"`
	Interval int  `json:"interval" yaml:"interval"`
}

type TimeoutConfig struct {
	HTTPClient      time.Duration `json:"httpClient" yaml:"http_client"`           // HTTP client timeout (default: 30s)
	ICEGathering    time.Duration `json:"iceGathering" yaml:"ice_gathering"`       // ICE gathering timeout (default: 5s)
	ICEConnection   time.Duration `json:"iceConnection" yaml:"ice_connection"`    // ICE connection timeout (default: 10s)
	ICECheckInterval time.Duration `json:"iceCheckInterval" yaml:"ice_check_interval"` // ICE check interval (default: 2s)
	ThroughputTicker time.Duration `json:"throughputTicker" yaml:"throughput_ticker"`   // Throughput stats ticker interval (default: 100ms)
	StopDelay       time.Duration `json:"stopDelay" yaml:"stop_delay"`              // Delay before stopping connections (default: 1s)
}

type Config struct {
	NodeID    string               `json:"nodeId" yaml:"node_id"`
	ICEConfig map[string]ICEConfig `json:"iceServers" yaml:"ice_servers"`
	Logging   LoggingConfig        `json:"logging" yaml:"logging"`
	Timer     TimerConfig          `json:"timer" yaml:"timer"`
	Api       ApiConfig            `json:"api" yaml:"api"`
	Timeouts  TimeoutConfig        `json:"timeouts" yaml:"timeouts"`

	WebRTCConfig webrtc.Configuration
	// TODO the following should be different for answerer and offerer sides
	OnICECandidate          func(*webrtc.ICECandidate)
	OnConnectionStateChange func(s webrtc.PeerConnectionState)

	// internal
	ServiceName string `yaml:"-"`
	Logger      *slog.Logger
	Registry    *prometheus.Registry
}

func mergeConfigs(c, responseConfig interface{}) {
	mergeStructs(reflect.ValueOf(c).Elem(), reflect.ValueOf(responseConfig).Elem())
}

func mergeStructs(cValue, respValue reflect.Value) {
	for i := 0; i < respValue.NumField(); i++ {
		respField := respValue.Field(i)
		cField := cValue.Field(i)

		if !respField.IsZero() {
			switch respField.Kind() {
			case reflect.Ptr:
				if !respField.IsNil() {
					if cField.IsNil() {
						cField.Set(reflect.New(cField.Type().Elem()))
					}
					mergeStructs(cField.Elem(), respField.Elem())
				}
			case reflect.Struct:
				mergeStructs(cField, respField)
			case reflect.Map:
				if cField.IsNil() {
					cField.Set(reflect.MakeMap(cField.Type()))
				}
				for _, key := range respField.MapKeys() {
					val := respField.MapIndex(key)
					cField.SetMapIndex(key, val)
				}
			default:
				cField.Set(respField)
			}
		}
	}
}

func NewConfig(confString string) (*Config, error) {
	c := &Config{
		ServiceName: "ICEPerf",
		Timeouts: TimeoutConfig{
			HTTPClient:      30 * time.Second,
			ICEGathering:    5 * time.Second,
			ICEConnection:   10 * time.Second,
			ICECheckInterval: 2 * time.Second,
			ThroughputTicker: 100 * time.Millisecond,
			StopDelay:       1 * time.Second,
		},
	}
	if confString != "" {
		if err := yaml.Unmarshal([]byte(confString), c); err != nil {
			return nil, err
		}
	}
	// Ensure defaults are set if not provided in config
	c.setTimeoutDefaults()
	return c, nil
}

func (c *Config) setTimeoutDefaults() {
	if c.Timeouts.HTTPClient == 0 {
		c.Timeouts.HTTPClient = 30 * time.Second
	}
	if c.Timeouts.ICEGathering == 0 {
		c.Timeouts.ICEGathering = 5 * time.Second
	}
	if c.Timeouts.ICEConnection == 0 {
		c.Timeouts.ICEConnection = 10 * time.Second
	}
	if c.Timeouts.ICECheckInterval == 0 {
		c.Timeouts.ICECheckInterval = 2 * time.Second
	}
	if c.Timeouts.ThroughputTicker == 0 {
		c.Timeouts.ThroughputTicker = 100 * time.Millisecond
	}
	if c.Timeouts.StopDelay == 0 {
		c.Timeouts.StopDelay = 1 * time.Second
	}
}

func (c *Config) UpdateConfigFromApi() error {
	c.setTimeoutDefaults()
	httpClient := &http.Client{
		Timeout: c.Timeouts.HTTPClient,
	}

	req, err := http.NewRequest("GET", c.Api.URI, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.Api.ApiKey)

	if err != nil {
		return err
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	//check the code of the response
	if res.StatusCode != 200 {
		return fmt.Errorf("API request failed with status %s: %w", res.Status, errors.New("non-200 status code"))
	}

	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	responseConfig := Config{}
	if err := json.Unmarshal([]byte(responseData), &responseConfig); err != nil {
		return fmt.Errorf("failed to unmarshal API response: %w", err)
	}

	//go and merge in values from the API into the config

	//lets just do the basics for now....
	//this needs a lot more work
	c.NodeID = responseConfig.NodeID
	c.ICEConfig = responseConfig.ICEConfig
	c.Logging = responseConfig.Logging
	// mergeConfigs(c, responseConfig)
	return nil
}
