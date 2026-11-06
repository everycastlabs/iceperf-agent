package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/nimbleape/iceperf-agent/config"
	"github.com/nimbleape/iceperf-agent/stats"
	"github.com/pion/stun/v2"
	"github.com/pion/webrtc/v4"
	"github.com/rs/xid"
	// "github.com/prometheus/client_golang/prometheus"
	// log "github.com/sirupsen/logrus"
)

type Client struct {
	ConnectionPair    *ConnectionPair
	OffererConnected  chan bool
	AnswererConnected chan bool
	close             chan struct{}
	Logger            *slog.Logger
	provider          string
	Stats             *stats.Stats
	config            *config.Config

	// Timing state (moved from package-level variables to prevent race conditions)
	startTime                     time.Time
	timeAnswererReceivedCandidate time.Time
	timeOffererReceivedCandidate  time.Time
	timeAnswererConnecting        time.Time
	timeAnswererConnected         time.Time
	timeOffererConnecting         time.Time
	timeOffererConnected          time.Time
}

func NewClient(config *config.Config, iceServerInfo *stun.URI, provider string, testRunId xid.ID, testRunStartedAt time.Time, doThroughputTest bool, close chan struct{}) (c *Client, err error) {
	return newClient(config, iceServerInfo, provider, testRunId, testRunStartedAt, doThroughputTest, close)
}

func newClient(cc *config.Config, iceServerInfo *stun.URI, provider string, testRunId xid.ID, testRunStartedAt time.Time, doThroughputTest bool, close chan struct{}) (*Client, error) {

	// Start timers
	startTime := time.Now()

	stats := stats.NewStats(testRunId.String(), testRunStartedAt)

	stats.SetProvider(provider)
	stats.SetScheme(iceServerInfo.Scheme.String())
	stats.SetProtocol(iceServerInfo.Proto.String())
	stats.SetPort(fmt.Sprintf("%d", iceServerInfo.Port))
	stats.SetNode(cc.NodeID)

	connectionPair, err := newConnectionPair(cc, iceServerInfo, provider, stats, doThroughputTest, close)

	if err != nil {
		return nil, err
	}

	c := &Client{
		ConnectionPair:    connectionPair,
		OffererConnected:  make(chan bool),
		AnswererConnected: make(chan bool),
		close:             make(chan struct{}),
		Logger:            cc.Logger,
		provider:          provider,
		Stats:             stats,
		config:            cc,
		startTime:         startTime,
	}

	if cc.OnICECandidate != nil {
		c.ConnectionPair.AnswerPC.OnICECandidate(cc.OnICECandidate)
		c.ConnectionPair.OfferPC.OnICECandidate(cc.OnICECandidate)
	} else {
		// Set ICE Candidate handler. As soon as a PeerConnection has gathered a candidate
		// send it to the other peer

		//if the test is a STUN test, then allow host, srflx and relay candidates
		c.ConnectionPair.AnswerPC.OnICECandidate(func(i *webrtc.ICECandidate) {
			if i != nil {
				if i.Typ == webrtc.ICECandidateTypeSrflx || i.Typ == webrtc.ICECandidateTypeRelay || (i.Typ == webrtc.ICECandidateTypeHost && (iceServerInfo.Scheme == stun.SchemeTypeSTUN || iceServerInfo.Scheme == stun.SchemeTypeSTUNS)) {
					stats.SetAnswererTimeToReceiveCandidate(float64(time.Since(c.startTime).Milliseconds()))
					c.timeAnswererReceivedCandidate = time.Now()
					c.ConnectionPair.LogAnswerer.Info("Answerer received candidate, sent over to other PC", "eventTime", c.timeAnswererReceivedCandidate,
						"timeSinceStartMs", time.Since(c.startTime).Milliseconds(),
						"candidateType", i.Typ,
						"relayAddress", i.RelatedAddress,
						"relayPort", i.RelatedPort)
					if err := c.ConnectionPair.OfferPC.AddICECandidate(i.ToJSON()); err != nil {
						c.ConnectionPair.LogOfferer.Error("failed to add ICE candidate to offerer", "error", err)
					}
				}
			}
		})

		// Set ICE Candidate handler. As soon as a PeerConnection has gathered a candidate
		// send it to the other peer
		c.ConnectionPair.OfferPC.OnICECandidate(func(i *webrtc.ICECandidate) {
			if i != nil {
				if i.Typ == webrtc.ICECandidateTypeSrflx || i.Typ == webrtc.ICECandidateTypeRelay {
					stats.SetOffererTimeToReceiveCandidate(float64(time.Since(c.startTime).Milliseconds()))
					c.timeOffererReceivedCandidate = time.Now()
					c.ConnectionPair.LogOfferer.Info("Offerer received candidate, sent over to other PC", "eventTime", c.timeOffererReceivedCandidate,
						"timeSinceStartMs", time.Since(c.startTime).Milliseconds(),
						"candidateType", i.Typ,
						"relayAddress", i.RelatedAddress,
						"relayPort", i.RelatedPort)
					if err := c.ConnectionPair.AnswerPC.AddICECandidate(i.ToJSON()); err != nil {
						c.ConnectionPair.LogAnswerer.Error("failed to add ICE candidate to answerer", "error", err)
					}
				}
			}
		})
	}

	if cc.OnConnectionStateChange != nil {
		c.ConnectionPair.OfferPC.OnConnectionStateChange(cc.OnConnectionStateChange)
		c.ConnectionPair.AnswerPC.OnConnectionStateChange(cc.OnConnectionStateChange)
	} else {
		// Set the handler for Peer connection state
		// This will notify you when the peer has connected/disconnected
		c.ConnectionPair.OfferPC.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
			c.ConnectionPair.LogOfferer.Info("Peer Connection State has changed", "eventTime", time.Now(),
				"peerConnState", s.String())

			switch s {
			case webrtc.PeerConnectionStateConnecting:
				c.timeOffererConnecting = time.Now()
				c.ConnectionPair.LogOfferer.Info("Offerer connecting", "eventTime", c.timeOffererConnecting,
					"timeSinceStartMs", time.Since(c.startTime).Milliseconds())
			case webrtc.PeerConnectionStateConnected:
				c.timeOffererConnected = time.Now()
				c.ConnectionPair.LogOfferer.Info("Offerer connected", "eventTime", c.timeOffererConnected, "timeSinceStartMs", time.Since(c.startTime).Milliseconds())
				//go and get the details about the ice pair
				//stats := c.ConnectionPair.OfferPC.GetStats()
				// connStats, ok := stats.GetConnectionStats(c.ConnectionPair.OfferPC)
				// if (ok) {
				// 	c.ConnectionPair.LogOfferer.WithFields(log.Fields{
				// 		"connStats": connStats,
				// 		"eventTime":        timeOffererConnected,
				// 		"timeSinceStartMs": time.Since(startTime).Milliseconds(),
				// 	}).Info("Offerer Stats")
				// }
				// find the active candidate pair
				// for k, v := range stats {
				// 	c.ConnectionPair.LogOfferer.WithFields(log.Fields{
				// 		"statsKey": k,
				// 		"statsValue": v,
				// 		"eventTime":        timeOffererConnected,
				// 		"timeSinceStartMs": time.Since(startTime).Milliseconds(),
				// 	}).Info("Offerer Stats")
				// }
				stats.SetTimeToConnectedState(time.Since(c.startTime).Milliseconds())
				c.OffererConnected <- true
			case webrtc.PeerConnectionStateFailed:
				c.ConnectionPair.LogOfferer.Error("Offerer connection failed", "eventTime", time.Now(), "timeSinceStartMs", time.Since(c.startTime).Milliseconds())
				close <- struct{}{}
				c.OffererConnected <- false
			case webrtc.PeerConnectionStateClosed:
				// PeerConnection was explicitly closed. This usually happens from a DTLS CloseNotify
				c.ConnectionPair.LogOfferer.Info("Offerer connection closed", "eventTime", time.Now(), "timeSinceStartMs", time.Since(c.startTime).Milliseconds())
				c.OffererConnected <- false
			}
		})

		// Set the handler for Peer connection state
		// This will notify you when the peer has connected/disconnected
		c.ConnectionPair.AnswerPC.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
			c.ConnectionPair.LogAnswerer.Info("Peer Connection State has changed", "eventTime", time.Now(), "peerConnState", s.String())

			switch s {
			case webrtc.PeerConnectionStateConnecting:
				c.timeAnswererConnecting = time.Now()
				c.ConnectionPair.LogAnswerer.Info("Answerer connecting", "eventTime", c.timeAnswererConnecting, "timeSinceStartMs", time.Since(c.startTime).Milliseconds())
			case webrtc.PeerConnectionStateConnected:
				c.timeAnswererConnected = time.Now()
				c.ConnectionPair.LogAnswerer.Info("Answerer connected", "eventTime", c.timeAnswererConnected, "timeSinceStartMs", time.Since(c.startTime).Milliseconds())
				c.AnswererConnected <- true
			case webrtc.PeerConnectionStateFailed:
				// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
				// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
				// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
				c.ConnectionPair.LogAnswerer.Error("Answerer connection failed", "eventTime", time.Now(), "timeSinceStartMs", time.Since(c.startTime).Milliseconds())
				c.AnswererConnected <- false
			case webrtc.PeerConnectionStateClosed:
				// PeerConnection was explicitly closed. This usually happens from a DTLS CloseNotify
				c.ConnectionPair.LogAnswerer.Info("Answerer connection closed", "eventTime", time.Now(), "timeSinceStartMs", time.Since(c.startTime).Milliseconds())
				c.AnswererConnected <- false
			}
		})
	}

	return c, nil
}

func (c *Client) Run() {
	go c.run()
}

func (c *Client) run() {
	offer, err := c.ConnectionPair.OfferPC.CreateOffer(nil)
	if err != nil {
		c.Logger.Error("failed to create offer", "error", err)
		return
	}
	if err := c.ConnectionPair.OfferPC.SetLocalDescription(offer); err != nil {
		c.Logger.Error("failed to set local description", "error", err)
		return
	}
	desc, err := json.Marshal(offer)
	if err != nil {
		c.Logger.Error("failed to marshal offer", "error", err)
		return
	}

	if err := c.ConnectionPair.setRemoteDescription(c.ConnectionPair.AnswerPC, desc); err != nil {
		c.Logger.Error("failed to set remote description on answerer", "error", err)
		return
	}

	answer, err := c.ConnectionPair.AnswerPC.CreateAnswer(nil)
	if err != nil {
		c.Logger.Error("failed to create answer", "error", err)
		return
	}
	if err := c.ConnectionPair.AnswerPC.SetLocalDescription(answer); err != nil {
		c.Logger.Error("failed to set local description on answerer", "error", err)
		return
	}
	desc2, err := json.Marshal(answer)
	if err != nil {
		c.Logger.Error("failed to marshal answer", "error", err)
		return
	}

	if err := c.ConnectionPair.setRemoteDescription(c.ConnectionPair.OfferPC, desc2); err != nil {
		c.Logger.Error("failed to set remote description on offerer", "error", err)
		return
	}

	// this is blocking
	c.close <- struct{}{}

	if err := c.Stop(); err != nil {
		c.Logger.Error("failed to stop client", "error", err)
	}
}

func (c *Client) Stop() error {
	c.Logger.Info("Stopping client...")

	if c.ConnectionPair.OfferDC != nil {
		c.ConnectionPair.OfferDC.Close()
	}

	time.Sleep(c.config.Timeouts.StopDelay)

	if err := c.ConnectionPair.OfferPC.Close(); err != nil {
		c.Logger.Error("cannot close c.ConnectionPair.OfferPC", "error", err)
		return err
	}

	if err := c.ConnectionPair.AnswerPC.Close(); err != nil {
		c.Logger.Error("cannot close c.ConnectionPair.AnswerPC", "error", err)
		return err
	}

	if c.config.Logging.API.Enabled {
		// Convert data to JSON
		c.Stats.CreateLabels()
		jsonData, err := json.Marshal(c.Stats)
		if err != nil {
			c.Logger.Error("failed to marshal stats to JSON", "error", err)
			return err
		}

		// Define the API endpoint
		apiEndpoint := c.config.Logging.API.URI

		// Create a new HTTP request
		req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			c.Logger.Error("failed to create HTTP request", "error", err, "endpoint", apiEndpoint)
			return err
		}

		// Set the appropriate headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+c.config.Logging.API.ApiKey)

		// Send the request using the HTTP client
		client := &http.Client{
			Timeout: c.config.Timeouts.HTTPClient,
		}
		resp, err := client.Do(req)
		if err != nil {
			c.Logger.Error("failed to send HTTP request", "error", err, "endpoint", apiEndpoint)
			return err
		}
		defer resp.Body.Close()

		// Check the response
		if resp.StatusCode == http.StatusCreated {
			c.Logger.Info("stats data sent successfully to API", "endpoint", apiEndpoint)
		} else {
			c.Logger.Warn("failed to send stats data to API", "statusCode", resp.StatusCode, "endpoint", apiEndpoint)
		}
	}
	j, err := c.Stats.ToJSON()
	if err != nil {
		c.Logger.Error("failed to convert stats to JSON", "error", err)
	} else {
		c.Logger.Info(j, "individual_test_completed", "true")
	}

	return nil
}
