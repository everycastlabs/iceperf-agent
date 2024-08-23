package config

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/pion/webrtc/v4"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"
)

type ICEConfig struct {
	Username     string           `yaml:"username,omitempty"`
	Password     string           `yaml:"password,omitempty"`
	ApiKey       string           `json:"apiKey,omitempty" yaml:"api_key,omitempty"`
	AccountSid   string           `yaml:"account_sid,omitempty"`
	RequestUrl   string           `json:"requestUrl,omitempty" yaml:"request_url,omitempty"`
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
	TurnToTurn   bool             `yaml:"turn_to_turn"`
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

type Config struct {
	NodeID    string               `json:"nodeId" yaml:"node_id"`
	ICEConfig map[string]ICEConfig `json:"iceServers" yaml:"ice_servers"`
	Logging   LoggingConfig        `json:"logging" yaml:"logging"`
	Timer     TimerConfig          `json:"timer" yaml:"timer"`
	Api       ApiConfig            `json:"api" yaml:"api"`

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
	}
	if confString != "" {
		if err := yaml.Unmarshal([]byte(confString), c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *Config) UpdateConfigFromApi() error {
	httpClient := &http.Client{}

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
		err = errors.New("error from our api " + res.Status)
		return err
	}

	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	responseConfig := Config{}
	json.Unmarshal([]byte(responseData), &responseConfig)

	//go and merge in values from the API into the config

	//lets just do the basics for now....
	//this needs a lot more work
	c.NodeID = responseConfig.NodeID
	c.ICEConfig = responseConfig.ICEConfig
	c.Logging = responseConfig.Logging
	// mergeConfigs(c, responseConfig)
	return nil
}
