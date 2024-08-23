package stats

import (
	"encoding/json"
	"time"
)

// Stats represents a statistics object
// Stats represents a statistics object
type Stats struct {
	TestRunID                              string            `json:"testRunID"`
	Labels                                 map[string]string `json:"labels"`
	AnswererTimeToReceiveCandidate         float64           `json:"answererTimeToReceiveCandidate"`
	OffererTimeToReceiveCandidate          float64           `json:"offererTimeToReceiveCandidate"`
	OffererDcBytesSentTotal                float64           `json:"offererDcBytesSentTotal"`
	OffererIceTransportBytesSentTotal      float64           `json:"offererIceTransportBytesSentTotal"`
	OffererIceTransportBytesReceivedTotal  float64           `json:"offererIceTransportBytesReceivedTotal"`
	AnswererDcBytesReceivedTotal           float64           `json:"answererDcBytesReceivedTotal"`
	AnswererIceTransportBytesReceivedTotal float64           `json:"answererIceTransportBytesReceivedTotal"`
	AnswererIceTransportBytesSentTotal     float64           `json:"answererIceTransportBytesSentTotal"`
	LatencyFirstPacket                     float64           `json:"latencyFirstPacket"`
	Throughput                             map[int64]float64 `json:"throughput"`
	InstantThroughput                      map[int64]float64 `json:"instantThroughput"`
	ThroughputMax                          float64           `json:"throughputMax"`
	TestRunStartedAt                       time.Time         `json:"testRunStartedAt"`
	Provider                               string            `json:"provider"`
	Scheme                                 string            `json:"scheme"`
	Protocol                               string            `json:"protocol"`
	Port                                   string            `json:"port"`
	Node                                   string            `json:"node"`
	TimeToConnectedState                   int64             `json:"timeToConnectedState"`
	Connected                              bool              `json:"connected"`
}

// NewStats creates a new Stats object with a given test run ID
func NewStats(testRunID string, testRunStartedAt time.Time) *Stats {
	s := &Stats{
		TestRunID:         testRunID,
		TestRunStartedAt:  testRunStartedAt,
		Throughput:        make(map[int64]float64), // Initialize the Throughput map
		InstantThroughput: make(map[int64]float64), // Initialize the Throughput map
		Connected:         false,
	}

	return s
}

func (s *Stats) SetTimeToConnectedState(t int64) {
	s.TimeToConnectedState = t
	s.Connected = true
}

func (s *Stats) SetProvider(st string) {
	s.Provider = st
}

func (s *Stats) SetScheme(st string) {
	s.Scheme = st
}

func (s *Stats) SetProtocol(st string) {
	s.Protocol = st
}

func (s *Stats) SetPort(st string) {
	s.Port = st
}

func (s *Stats) SetNode(st string) {
	s.Node = st
}

func (s *Stats) SetOffererTimeToReceiveCandidate(o float64) {
	s.OffererTimeToReceiveCandidate = o
}

func (s *Stats) SetAnswererTimeToReceiveCandidate(o float64) {
	s.AnswererTimeToReceiveCandidate = o
}

func (s *Stats) SetOffererDcBytesSentTotal(d float64) {
	s.OffererDcBytesSentTotal = d
}

func (s *Stats) SetOffererIceTransportBytesSentTotal(io float64) {
	s.OffererIceTransportBytesSentTotal = io
}

func (s *Stats) SetOffererIceTransportBytesReceivedTotal(io float64) {
	s.OffererIceTransportBytesReceivedTotal = io
}

func (s *Stats) SetAnswererDcBytesReceivedTotal(a float64) {
	s.AnswererDcBytesReceivedTotal = a
}

func (s *Stats) SetAnswererIceTransportBytesReceivedTotal(ia float64) {
	s.AnswererIceTransportBytesReceivedTotal = ia
}

func (s *Stats) SetAnswererIceTransportBytesSentTotal(ia float64) {
	s.AnswererIceTransportBytesSentTotal = ia
}

func (s *Stats) SetLatencyFirstPacket(l float64) {
	s.LatencyFirstPacket = l
}

func (s *Stats) setThroughputMax(l float64) {
	if l < s.ThroughputMax {
		// New value is lower, don't set it
		return
	}
	s.ThroughputMax = l
}

func (s *Stats) AddThroughput(tp int64, v float64, v2 float64) {
	s.setThroughputMax(v2)
	if _, ok := s.Throughput[tp]; !ok {
		s.Throughput[tp] = v
	} else {
		s.Throughput[tp] += v
	}

	if _, ok := s.InstantThroughput[tp]; !ok {
		s.InstantThroughput[tp] = v2
	} else {
		s.InstantThroughput[tp] += v2
	}
}

func (s *Stats) CreateLabels() {
	s.Labels = map[string]string{
		"provider": s.Provider,
		"scheme":   s.Scheme,
		"protocol": s.Protocol,
		"port":     s.Port,
		"location": s.Node,
	}
}

// ToJSON returns the stats object as a JSON string
func (s *Stats) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(s)

	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
