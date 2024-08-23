package client

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nimbleape/iceperf-agent/config"
	"github.com/nimbleape/iceperf-agent/stats"
	"github.com/nimbleape/iceperf-agent/util"
	"github.com/pion/stun/v2"
	"github.com/pion/webrtc/v4"
)

type PC struct {
	pc *webrtc.PeerConnection
}

func (pc *PC) Stop() {

}

type ConnectionPair struct {
	OfferPC                 *webrtc.PeerConnection
	OfferDC                 *webrtc.DataChannel
	AnswerPC                *webrtc.PeerConnection
	LogOfferer              *slog.Logger
	LogAnswerer             *slog.Logger
	config                  *config.Config
	sentInitialMessageViaDC time.Time
	iceServerInfo           *stun.URI
	provider                string
	stats                   *stats.Stats
	doThroughputTest        bool
	closeChan               chan struct{}
}

func NewConnectionPair(config *config.Config, iceServerInfo *stun.URI, provider string, stats *stats.Stats, doThroughputTest bool, closeChan chan struct{}) (c *ConnectionPair, err error) {
	return newConnectionPair(config, iceServerInfo, provider, stats, doThroughputTest, closeChan)
}

func newConnectionPair(cc *config.Config, iceServerInfo *stun.URI, provider string, stats *stats.Stats, doThroughputTest bool, closeChan chan struct{}) (*ConnectionPair, error) {
	logOfferer := cc.Logger.With("peer", "Offerer")
	logAnswerer := cc.Logger.With("peer", "Answerer")

	cp := &ConnectionPair{
		config:           cc,
		LogOfferer:       logOfferer,
		LogAnswerer:      logAnswerer,
		iceServerInfo:    iceServerInfo,
		provider:         provider,
		stats:            stats,
		doThroughputTest: doThroughputTest,
		closeChan:        closeChan,
	}

	config := webrtc.Configuration{}

	if cc.WebRTCConfig.ICEServers != nil {
		config.ICEServers = cc.WebRTCConfig.ICEServers
	}

	config.ICETransportPolicy = cc.WebRTCConfig.ICETransportPolicy
	config.SDPSemantics = webrtc.SDPSemanticsUnifiedPlanWithFallback

	//we only want offerer to force turn (if we are)
	cp.createOfferer(config)

	// think we want to leave the answerer without any ice servers so we only get the host candidates.... I think
	// to get the tests working I'm passing the turn server into both....
	// but I don't think that should be required
	cp.createAnswerer(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})

	return cp, nil
}

func (cp *ConnectionPair) setRemoteDescription(pc *webrtc.PeerConnection, sdp []byte) {
	var desc webrtc.SessionDescription
	err := json.Unmarshal(sdp, &desc)
	util.Check(err)

	// Apply the desc as the remote description
	err = pc.SetRemoteDescription(desc)
	util.Check(err)
}

func (cp *ConnectionPair) createOfferer(config webrtc.Configuration) {
	// Create a new PeerConnection
	settingEngine := webrtc.SettingEngine{}
	settingEngine.SetICETimeouts(5*time.Second, 10*time.Second, 2*time.Second)
	api := webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine))

	pc, err := api.NewPeerConnection(config)
	util.Check(err)

	buf := make([]byte, 1024)

	ordered := false
	maxRetransmits := uint16(0)
	hasSentData := false

	options := &webrtc.DataChannelInit{
		Ordered:        &ordered,
		MaxRetransmits: &maxRetransmits,
	}

	sendMoreCh := make(chan struct{}, 1)

	// Create a datachannel with label 'data'
	dc, err := pc.CreateDataChannel("data", options)
	util.Check(err)

	cp.OfferDC = dc

	if cp.iceServerInfo.Scheme == stun.SchemeTypeTURN || cp.iceServerInfo.Scheme == stun.SchemeTypeTURNS {

		// labels := map[string]string{
		// 	"provider": cp.provider,
		// 	"scheme":   cp.iceServerInfo.Scheme.String(),
		// 	"protocol": cp.iceServerInfo.Proto.String(),
		// 	"port":     fmt.Sprintf("%d", cp.iceServerInfo.Port),
		// }

		// Register channel opening handling
		dc.OnOpen(func() {

			stats := pc.GetStats()
			iceTransportStats := stats["iceTransport"].(webrtc.TransportStats)
			// for k, v := range stats {
			cp.LogOfferer.Info("Offerer Stats", "iceTransportStats", iceTransportStats.BytesReceived)
			//}

			cp.LogOfferer.Info("OnOpen: Start sending a series of 1024-byte packets as fast as it can", "dataChannelLabel", dc.Label(),
				"dataChannelId", dc.ID(),
			)

			for {
				if !hasSentData {
					cp.sentInitialMessageViaDC = time.Now()
					hasSentData = true
				}
				err2 := dc.Send(buf)
				if err2 != nil {
					break
				}

				if dc.BufferedAmount() > maxBufferedAmount {
					// Wait until the bufferedAmount becomes lower than the threshold
					<-sendMoreCh
				}
			}
		})

		// Set bufferedAmountLowThreshold so that we can get notified when
		// we can send more
		dc.SetBufferedAmountLowThreshold(bufferedAmountLowThreshold)

		// This callback is made when the current bufferedAmount becomes lower than the threshold
		dc.OnBufferedAmountLow(func() {
			// Make sure to not block this channel or perform long running operations in this callback
			// This callback is executed by pion/sctp. If this callback is blocking it will stop operations
			if cp.doThroughputTest {
				select {
				case sendMoreCh <- struct{}{}:
				default:
				}
			} else {
				//a noop
			}
		})

		dc.OnClose(func() {

			dcBytesSentTotal, _, iceTransportSentBytesTotal, iceTransportReceivedBytesTotal, _ := getBytesStats(pc, dc)

			cp.stats.SetOffererDcBytesSentTotal(float64(dcBytesSentTotal))
			cp.stats.SetOffererIceTransportBytesSentTotal(float64(iceTransportSentBytesTotal))
			cp.stats.SetOffererIceTransportBytesReceivedTotal(float64(iceTransportReceivedBytesTotal))

			cp.LogOfferer.Info("Sent total", "dcSentBytesTotal", dcBytesSentTotal,
				"cpSentBytesTotal", iceTransportSentBytesTotal)
		})
	}
	cp.OfferPC = pc
}

func (cp *ConnectionPair) createAnswerer(config webrtc.Configuration) {

	// settingEngine := webrtc.SettingEngine{}
	// settingEngine.SetICETimeouts(5, 5, 2)
	// api := webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine))
	// Create a new PeerConnection
	pc, err := webrtc.NewPeerConnection(config)
	util.Check(err)

	if cp.iceServerInfo.Scheme == stun.SchemeTypeTURN || cp.iceServerInfo.Scheme == stun.SchemeTypeTURNS {

		// labels := map[string]string{
		// 	"provider": cp.provider,
		// 	"scheme":   cp.iceServerInfo.Scheme.String(),
		// 	"protocol": cp.iceServerInfo.Proto.String(),
		// 	"port":     fmt.Sprintf("%d", cp.iceServerInfo.Port),
		// }

		pc.OnDataChannel(func(dc *webrtc.DataChannel) {
			var totalBytesReceived uint64

			hasReceivedData := false

			// Register channel opening handling
			dc.OnOpen(func() {

				cp.LogAnswerer.Info("OnOpen: Start receiving data", "dataChannelLabel", dc.Label(),
					"dataChannelId", dc.ID())

				since := time.Now()

				lastTotalBytesReceived := uint64(0)
				// Start printing out the observed throughput
				for range time.NewTicker(100 * time.Millisecond).C {
					//check if this pc is closed and break out
					if pc.ConnectionState() != webrtc.PeerConnectionStateConnected {
						break
					}
					_, totalBytesReceivedTmp, _, _, ok := getBytesStats(pc, dc)
					if ok {
						totalBytesReceived = totalBytesReceivedTmp
						// cp.LogAnswerer.Info("Received Bytes So Far", "dcReceivedBytes", totalBytesReceivedTmp,
						// 	"cpReceivedBytes", cpTotalBytesReceivedTmp)
					}

					bytesLastTicker := totalBytesReceived - lastTotalBytesReceived

					bps := 8 * float64(bytesLastTicker) * 10
					lastTotalBytesReceived = totalBytesReceivedTmp

					averageBps := 8 * float64(totalBytesReceived) / float64(time.Since(since).Seconds())
					// bps := float64(atomic.LoadUint64(&totalBytesReceived)*8) / time.Since(since).Seconds()
					cp.LogAnswerer.Info("On ticker: Calculated throughput", "bytesLastTicker", bytesLastTicker, "throughput", bps/1024/1024, "avgthroughput", averageBps/1024/1024, "eventTime", time.Now())
					if cp.doThroughputTest {
						cp.stats.AddThroughput(time.Since(since).Milliseconds(), averageBps/1024/1024, bps/1024/1024)
					}
				}

				bps := 8 * float64(totalBytesReceived) / float64(time.Since(since).Seconds())
				// bps := float64(atomic.LoadUint64(&totalBytesReceived)*8) / time.Since(since).Seconds()
				cp.LogAnswerer.Info("On ticker: Calculated throughput", "throughput", bps/1024/1024,
					"eventTime", time.Now(),
					"timeSinceStartMs", time.Since(since).Milliseconds())
				if cp.doThroughputTest {
					cp.stats.AddThroughput(time.Since(since).Milliseconds(), bps/1024/1024, 0)
				}
			})

			// Register the OnMessage to handle incoming messages
			dc.OnMessage(func(dcMsg webrtc.DataChannelMessage) {

				if !hasReceivedData {
					cp.stats.SetLatencyFirstPacket(float64(time.Since(cp.sentInitialMessageViaDC).Milliseconds()))
					cp.LogAnswerer.Info("Received first Packet", "latencyFirstPacketInMs", time.Since(cp.sentInitialMessageViaDC).Milliseconds())
					hasReceivedData = true
				}
				if !cp.doThroughputTest {
					cp.LogAnswerer.Info("Sending to close")
					cp.closeChan <- struct{}{}
				}
			})

			dc.OnClose(func() {

				_, dcBytesReceivedTotal, iceTransportBytesReceivedTotal, iceTransportBytesSentTotal, _ := getBytesStats(pc, dc)

				cp.stats.SetAnswererDcBytesReceivedTotal(float64(dcBytesReceivedTotal))
				cp.stats.SetAnswererIceTransportBytesReceivedTotal(float64(iceTransportBytesReceivedTotal))
				cp.stats.SetAnswererIceTransportBytesSentTotal(float64(iceTransportBytesSentTotal))

				cp.LogAnswerer.Info("Received total", "dcReceivedBytesTotal", dcBytesReceivedTotal,
					"iceTransportReceivedBytesTotal", iceTransportBytesReceivedTotal)
			})
		})
	}

	cp.AnswerPC = pc
}

func getBytesStats(pc *webrtc.PeerConnection, dc *webrtc.DataChannel) (uint64, uint64, uint64, uint64, bool) {

	stats := pc.GetStats()

	// for _, report := range stats {
	// 	//if candidatePairStats, ok := report.(webrtc.ICECandidatePairStats); ok {
	// 	// Check if this candidate pair is the selected one
	// 	// if candidatePairStats.Nominated {
	// 	// fmt.Printf("WebRTC Stat: %+v\n", report)
	// 	//			}
	// 	//}
	// }

	dcStats, ok := stats.GetDataChannelStats(dc)
	if !ok {
		return 0, 0, 0, 0, ok
	}

	iceTransportStats := stats["iceTransport"].(webrtc.TransportStats)

	return dcStats.BytesSent, dcStats.BytesReceived, iceTransportStats.BytesSent, iceTransportStats.BytesReceived, ok
}
