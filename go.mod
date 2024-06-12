module github.com/nimbleape/iceperf-agent

go 1.21.6

require (
	github.com/alecthomas/assert/v2 v2.5.0
	github.com/joho/godotenv v1.5.1
	github.com/magnetde/slog-loki v0.1.4
	github.com/pion/stun/v2 v2.0.0
	github.com/pion/webrtc/v4 v4.0.0-beta.19
	github.com/prometheus/client_golang v1.19.1
	github.com/rs/xid v1.5.0
	github.com/samber/slog-multi v1.0.3
	github.com/urfave/cli/v2 v2.27.1
	gopkg.in/yaml.v3 v3.0.1
)

// replace github.com/pion/webrtc/v4 => ../pion-webrtc

require (
	github.com/alecthomas/repr v0.3.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/castai/promwrite v0.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hexops/gotextdiff v1.0.3 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pion/datachannel v1.5.6 // indirect
	github.com/pion/dtls/v2 v2.2.11 // indirect
	github.com/pion/ice/v3 v3.0.7 // indirect
	github.com/pion/interceptor v0.1.29 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns/v2 v2.0.7 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.14 // indirect
	github.com/pion/rtp v1.8.6 // indirect
	github.com/pion/sctp v1.8.16 // indirect
	github.com/pion/sdp/v3 v3.0.9 // indirect
	github.com/pion/srtp/v3 v3.0.1 // indirect
	github.com/pion/transport/v2 v2.2.4 // indirect
	github.com/pion/transport/v3 v3.0.2 // indirect
	github.com/pion/turn/v3 v3.0.3 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/prometheus/prometheus v0.40.3 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/samber/lo v1.38.1 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)
