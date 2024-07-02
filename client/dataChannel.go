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
}

func NewConnectionPair(config *config.Config, iceServerInfo *stun.URI, provider string, stats *stats.Stats) (c *ConnectionPair, err error) {
	return newConnectionPair(config, iceServerInfo, provider, stats)
}

func newConnectionPair(cc *config.Config, iceServerInfo *stun.URI, provider string, stats *stats.Stats) (*ConnectionPair, error) {

	logOfferer := cc.Logger.With("peer", "Offerer")
	logAnswerer := cc.Logger.With("peer", "Answerer")

	cp := &ConnectionPair{
		config:        cc,
		LogOfferer:    logOfferer,
		LogAnswerer:   logAnswerer,
		iceServerInfo: iceServerInfo,
		provider:      provider,
		stats:         stats,
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
	pc, err := webrtc.NewPeerConnection(config)
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
			select {
			case sendMoreCh <- struct{}{}:
			default:
			}
		})

		dc.OnClose(func() {

			dcBytesSentTotal, iceTransportSentBytesTotal, _ := getBytesSent(pc, dc)

			cp.stats.SetOffererDcBytesSentTotal(float64(dcBytesSentTotal))
			cp.stats.SetOffererIceTransportBytesSentTotal(float64(iceTransportSentBytesTotal))

			cp.LogOfferer.Info("Sent total", "dcSentBytesTotal", dcBytesSentTotal,
				"cpSentBytesTotal", iceTransportSentBytesTotal)
		})
	}
	cp.OfferPC = pc
}

func (cp *ConnectionPair) createAnswerer(config webrtc.Configuration) {
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

				// Start printing out the observed throughput
				for range time.NewTicker(100 * time.Millisecond).C {
					//check if this pc is closed and break out
					if pc.ConnectionState() != webrtc.PeerConnectionStateConnected {
						break
					}
					bps := 8 * float64(totalBytesReceived) / float64(time.Since(since).Seconds())
					// bps := float64(atomic.LoadUint64(&totalBytesReceived)*8) / time.Since(since).Seconds()
					cp.LogAnswerer.Info("On ticker: Calculated throughput", "throughput", bps/1024/1024,
						"eventTime", time.Now())

					cp.stats.AddThroughput(time.Since(since).Milliseconds(), bps/1024/1024)
				}
				bps := 8 * float64(totalBytesReceived) / float64(time.Since(since).Seconds())
				// bps := float64(atomic.LoadUint64(&totalBytesReceived)*8) / time.Since(since).Seconds()
				cp.LogAnswerer.Info("On ticker: Calculated throughput", "throughput", bps/1024/1024,
					"eventTime", time.Now(),
					"timeSinceStartMs", time.Since(since).Milliseconds())
				cp.stats.AddThroughput(time.Since(since).Milliseconds(), bps/1024/1024)
			})

			// Register the OnMessage to handle incoming messages
			dc.OnMessage(func(dcMsg webrtc.DataChannelMessage) {

				if !hasReceivedData {
					cp.stats.SetLatencyFirstPacket(float64(time.Since(cp.sentInitialMessageViaDC).Milliseconds()))
					cp.LogAnswerer.Info("Received first Packet", "latencyFirstPacketInMs", time.Since(cp.sentInitialMessageViaDC).Milliseconds())
					hasReceivedData = true
				}
				totalBytesReceivedTmp, _, ok := getBytesReceived(pc, dc)
				if ok {
					totalBytesReceived = totalBytesReceivedTmp
					// cp.LogAnswerer.Info("Received Bytes So Far", "dcReceivedBytes", totalBytesReceivedTmp,
					// 	"cpReceivedBytes", cpTotalBytesReceivedTmp)
				}
			})

			dc.OnClose(func() {

				dcBytesReceivedTotal, iceTransportBytesReceivedTotal, _ := getBytesReceived(pc, dc)

				cp.stats.SetAnswererDcBytesReceivedTotal(float64(dcBytesReceivedTotal))
				cp.stats.SetAnswererIceTransportBytesReceivedTotal(float64(iceTransportBytesReceivedTotal))

				cp.LogAnswerer.Info("Received total", "dcReceivedBytesTotal", dcBytesReceivedTotal,
					"iceTransportReceivedBytesTotal", iceTransportBytesReceivedTotal)
			})
		})
	}

	cp.AnswerPC = pc
}

func getBytesReceived(pc *webrtc.PeerConnection, dc *webrtc.DataChannel) (uint64, uint64, bool) {
	stats := pc.GetStats()

	dcStats, ok := stats.GetDataChannelStats(dc)
	if !ok {
		return 0, 0, ok
	}

	iceTransportStats := stats["iceTransport"].(webrtc.TransportStats)

	return dcStats.BytesReceived, iceTransportStats.BytesReceived, ok
}

func getBytesSent(pc *webrtc.PeerConnection, dc *webrtc.DataChannel) (uint64, uint64, bool) {
	stats := pc.GetStats()

	dcStats, ok := stats.GetDataChannelStats(dc)
	if !ok {
		return 0, 0, ok
	}

	iceTransportStats := stats["iceTransport"].(webrtc.TransportStats)

	return dcStats.BytesSent, iceTransportStats.BytesSent, ok
}
