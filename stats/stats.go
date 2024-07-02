package stats

import (
	"encoding/json"
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
	AnswererDcBytesReceivedTotal           float64           `json:"answererDcBytesReceivedTotal"`
	AnswererIceTransportBytesReceivedTotal float64           `json:"answererIceTransportBytesReceivedTotal"`
	LatencyFirstPacket                     float64           `json:"latencyFirstPacket"`
	Throughput                             map[int64]float64 `json:"throughput"`
	ThroughputMax                          float64           `json:"throughputMax"`
}

// NewStats creates a new Stats object with a given test run ID
func NewStats(testRunID string, labels map[string]string) *Stats {
	return &Stats{
		TestRunID:  testRunID,
		Labels:     labels,
		Throughput: make(map[int64]float64), // Initialize the Throughput map
	}
}

func (s *Stats) SetAnswererTimeToReceiveCandidate(c float64) {
	s.AnswererTimeToReceiveCandidate = c
}

func (s *Stats) SetOffererTimeToReceiveCandidate(o float64) {
	s.OffererTimeToReceiveCandidate = o
}

func (s *Stats) SetOffererDcBytesSentTotal(d float64) {
	s.OffererDcBytesSentTotal = d
}

func (s *Stats) SetOffererIceTransportBytesSentTotal(io float64) {
	s.OffererIceTransportBytesSentTotal = io
}

func (s *Stats) SetAnswererDcBytesReceivedTotal(a float64) {
	s.AnswererDcBytesReceivedTotal = a
}

func (s *Stats) SetAnswererIceTransportBytesReceivedTotal(ia float64) {
	s.AnswererIceTransportBytesReceivedTotal = ia
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

func (s *Stats) AddThroughput(tp int64, v float64) {
	s.setThroughputMax(v)
	if _, ok := s.Throughput[tp]; !ok {
		s.Throughput[tp] = v
	} else {
		s.Throughput[tp] += v
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
