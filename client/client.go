package client

import (
	"encoding/json"
	"time"

	"github.com/nimbleape/iceperf-agent/config"
	"github.com/nimbleape/iceperf-agent/util"
	"github.com/pion/stun/v2"
	"github.com/pion/webrtc/v4"
	log "github.com/sirupsen/logrus"
)

const (
	bufferedAmountLowThreshold uint64 = 512 * 1024  // 512 KB
	maxBufferedAmount          uint64 = 1024 * 1024 // 1 MB
)

var (
	startTime                     time.Time
	timeAnswererReceivedCandidate time.Time
	timeOffererReceivedCandidate  time.Time
	timeAnswererConnecting        time.Time
	timeAnswererConnected         time.Time
	timeOffererConnecting         time.Time
	timeOffererConnected          time.Time
)

type Client struct {
	ConnectionPair    *ConnectionPair
	OffererConnected  chan bool
	AnswererConnected chan bool
	close             chan struct{}
}

func NewClient(config *config.Config, iceServerInfo *stun.URI) (c *Client, err error) {
	return newClient(config, iceServerInfo)
}

func newClient(cc *config.Config, iceServerInfo *stun.URI) (*Client, error) {

	// Start timers
	startTime = time.Now()

	connectionPair, err := newConnectionPair(cc, iceServerInfo)

	if err != nil {
		return nil, err
	}

	c := &Client{
		ConnectionPair:    connectionPair,
		OffererConnected:  make(chan bool),
		AnswererConnected: make(chan bool),
		close:             make(chan struct{}),
	}

	if cc.OnICECandidate != nil {
		c.ConnectionPair.AnswerPC.OnICECandidate(cc.OnICECandidate)
		c.ConnectionPair.OfferPC.OnICECandidate(cc.OnICECandidate)
	} else {
		// Set ICE Candidate handler. As soon as a PeerConnection has gathered a candidate
		// send it to the other peer
		c.ConnectionPair.AnswerPC.OnICECandidate(func(i *webrtc.ICECandidate) {
			if i != nil {
				if i.Typ == webrtc.ICECandidateTypeSrflx || i.Typ == webrtc.ICECandidateTypeRelay {
					timeAnswererReceivedCandidate = time.Now()
					c.ConnectionPair.LogAnswerer.WithFields(log.Fields{
						"eventTime":        timeAnswererReceivedCandidate,
						"timeSinceStartMs": time.Since(startTime).Milliseconds(),
						"candidateType":    i.Typ,
						"relayAddress":     i.RelatedAddress,
						"relayPort":        i.RelatedPort,
					}).Info("Answerer received candidate, sent over to other PC")
					util.Check(c.ConnectionPair.OfferPC.AddICECandidate(i.ToJSON()))
				}
			}
		})

		// Set ICE Candidate handler. As soon as a PeerConnection has gathered a candidate
		// send it to the other peer
		c.ConnectionPair.OfferPC.OnICECandidate(func(i *webrtc.ICECandidate) {
			if i != nil {
				timeOffererReceivedCandidate = time.Now()
				c.ConnectionPair.LogOfferer.WithFields(log.Fields{
					"eventTime":        timeOffererReceivedCandidate,
					"timeSinceStartMs": time.Since(startTime).Milliseconds(),
					"candidateType":    i.Typ,
					"relayAddress":     i.RelatedAddress,
					"relayPort":        i.RelatedPort,
				}).Info("Offerer received candidate, sent over to other PC")
				util.Check(c.ConnectionPair.AnswerPC.AddICECandidate(i.ToJSON()))
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
			c.ConnectionPair.LogOfferer.WithFields(log.Fields{
				"eventTime":     time.Now(),
				"peerConnState": s.String(),
			}).Info("Peer Connection State has changed")

			switch s {
			case webrtc.PeerConnectionStateConnecting:
				timeOffererConnecting = time.Now()
				c.ConnectionPair.LogOfferer.WithFields(log.Fields{
					"eventTime":        timeOffererConnecting,
					"timeSinceStartMs": time.Since(startTime).Milliseconds(),
				}).Info("Offerer connecting")
			case webrtc.PeerConnectionStateConnected:
				timeOffererConnected = time.Now()
				c.ConnectionPair.LogOfferer.WithFields(log.Fields{
					"eventTime":        timeOffererConnected,
					"timeSinceStartMs": time.Since(startTime).Milliseconds(),
				}).Info("Offerer connected")
				c.OffererConnected <- true
			case webrtc.PeerConnectionStateFailed:
				// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
				// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
				// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
				c.ConnectionPair.LogOfferer.WithFields(log.Fields{
					"eventTime":        time.Now(),
					"timeSinceStartMs": time.Since(startTime).Milliseconds(),
				}).Error("Offerer connection failed")
				c.OffererConnected <- false
			case webrtc.PeerConnectionStateClosed:
				// PeerConnection was explicitly closed. This usually happens from a DTLS CloseNotify
				c.ConnectionPair.LogOfferer.WithFields(log.Fields{
					"eventTime":        time.Now(),
					"timeSinceStartMs": time.Since(startTime).Milliseconds(),
				}).Info("Offerer connection closed")
				c.OffererConnected <- false
			}
		})

		// Set the handler for Peer connection state
		// This will notify you when the peer has connected/disconnected
		c.ConnectionPair.AnswerPC.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
			c.ConnectionPair.LogAnswerer.WithFields(log.Fields{
				"eventTime":     time.Now(),
				"peerConnState": s.String(),
			}).Info("Peer Connection State has changed")

			switch s {
			case webrtc.PeerConnectionStateConnecting:
				timeAnswererConnecting = time.Now()
				c.ConnectionPair.LogAnswerer.WithFields(log.Fields{
					"eventTime":        timeAnswererConnecting,
					"timeSinceStartMs": time.Since(startTime).Milliseconds(),
				}).Info("Answerer connecting")
			case webrtc.PeerConnectionStateConnected:
				timeAnswererConnected = time.Now()
				c.ConnectionPair.LogAnswerer.WithFields(log.Fields{
					"eventTime":        timeAnswererConnected,
					"timeSinceStartMs": time.Since(startTime).Milliseconds(),
				}).Info("Answerer connected")
				c.AnswererConnected <- true
			case webrtc.PeerConnectionStateFailed:
				// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
				// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
				// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
				c.ConnectionPair.LogAnswerer.WithFields(log.Fields{
					"eventTime":        time.Now(),
					"timeSinceStartMs": time.Since(startTime).Milliseconds(),
				}).Error("Answerer connection failed")
				c.AnswererConnected <- false
			case webrtc.PeerConnectionStateClosed:
				// PeerConnection was explicitly closed. This usually happens from a DTLS CloseNotify
				c.ConnectionPair.LogAnswerer.WithFields(log.Fields{
					"eventTime":        time.Now(),
					"timeSinceStartMs": time.Since(startTime).Milliseconds(),
				}).Info("Answerer connection closed")
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
	util.Check(err)
	util.Check(c.ConnectionPair.OfferPC.SetLocalDescription(offer))
	desc, err := json.Marshal(offer)
	util.Check(err)

	c.ConnectionPair.setRemoteDescription(c.ConnectionPair.AnswerPC, desc)

	answer, err := c.ConnectionPair.AnswerPC.CreateAnswer(nil)
	util.Check(err)
	util.Check(c.ConnectionPair.AnswerPC.SetLocalDescription(answer))
	desc2, err := json.Marshal(answer)
	util.Check(err)

	c.ConnectionPair.setRemoteDescription(c.ConnectionPair.OfferPC, desc2)

	// this is blocking
	c.close <- struct{}{}
	util.Check(c.Stop())
}

func (c *Client) Stop() error {
	log.Info("Stopping client...")
	if err := c.ConnectionPair.OfferPC.Close(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("cannot close c.ConnectionPair.OfferPC")
		return err
	}

	if err := c.ConnectionPair.AnswerPC.Close(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("cannot close c.ConnectionPair.AnswerPC")
		return err
	}

	return nil
}
