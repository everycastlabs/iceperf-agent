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
	AnswererDcBytesReceivedTotal           float64           `json:"answererDcBytesReceivedTotal"`
	AnswererIceTransportBytesReceivedTotal float64           `json:"answererIceTransportBytesReceivedTotal"`
	LatencyFirstPacket                     float64           `json:"latencyFirstPacket"`
	Throughput                             map[int64]float64 `json:"throughput"`
	ThroughputMax                          float64           `json:"throughputMax"`
	TestRunStartedAt                       time.Time         `json:"testRunStartedAt"`
	Provider                               string            `json:"provider"`
	Scheme                                 string            `json:"scheme"`
	Protocol                               string            `json:"protocol"`
	Port                                   string            `json:"port"`
	Location                               string            `json:"location"`
}

// NewStats creates a new Stats object with a given test run ID
func NewStats(testRunID string, testRunStartedAt time.Time) *Stats {
	s := &Stats{
		TestRunID:        testRunID,
		TestRunStartedAt: testRunStartedAt,
		Throughput:       make(map[int64]float64), // Initialize the Throughput map
	}

	return s
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

func (s *Stats) SetLocation(st string) {
	s.Location = st
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

	s.Labels = map[string]string{
		"provider": s.Provider,
		"scheme":   s.Scheme,
		"protocol": s.Protocol,
		"port":     s.Port,
		"location": s.Location,
	}

	jsonBytes, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
