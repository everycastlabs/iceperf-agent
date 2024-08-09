// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/nimbleape/iceperf-agent/client"
	"github.com/nimbleape/iceperf-agent/config"
	"github.com/nimbleape/iceperf-agent/stats"
	"github.com/nimbleape/iceperf-agent/version"
	"github.com/pion/stun/v2"
	"github.com/pion/webrtc/v4"
	"github.com/prometheus/client_golang/prometheus"

	// "github.com/prometheus/client_golang/prometheus/push"

	"github.com/rs/xid"

	// slogloki "github.com/samber/slog-loki/v3"

	slogmulti "github.com/samber/slog-multi"
	"github.com/urfave/cli/v2"

	// "github.com/grafana/loki-client-go/loki"
	loki "github.com/magnetde/slog-loki"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

func main() {
	app := &cli.App{
		Name:        "ICEPerf",
		Usage:       "ICE Servers performance tests",
		Version:     version.Version,
		Description: "Run ICE Servers performance tests and report results",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "ICEPerf yaml config file",
			},
			&cli.StringFlag{
				Name:    "api-uri",
				Aliases: []string{"a"},
				Usage:   "API URI",
			},
			&cli.StringFlag{
				Name:    "api-key",
				Aliases: []string{"k"},
				Usage:   "API Key",
			},
			&cli.BoolFlag{
				Name:    "timer",
				Aliases: []string{"t"},
				Value:   false,
				Usage:   "Enable Timer Mode",
			},
		},
		Action: runService,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func runService(ctx *cli.Context) error {
	config, err := getConfig(ctx)
	if err != nil {
		fmt.Println("Error loading config")
		return err
	}

	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelError)

	// Configure the logger

	var logg *slog.Logger

	var loggingLevel slog.Level

	switch config.Logging.Level {
	case "debug":
		loggingLevel = slog.LevelDebug
	case "info":
		loggingLevel = slog.LevelInfo
	case "error":
		loggingLevel = slog.LevelError
	default:
		loggingLevel = slog.LevelInfo
	}

	if config.Logging.Loki.Enabled {

		// config, _ := loki.NewDefaultConfig(config.Logging.Loki.URL)
		// // config.TenantID = "xyz"
		// client, _ := loki.New(config)

		lokiHandler := loki.NewHandler(
			config.Logging.Loki.URL,
			loki.WithLabelsEnabled(loki.LabelAll...),
			loki.WithHandler(func(w io.Writer) slog.Handler {
				return slog.NewJSONHandler(w, &slog.HandlerOptions{
					Level: loggingLevel,
				})
			}))

		logg = slog.New(
			slogmulti.Fanout(
				slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
					Level: loggingLevel,
				}),
				// slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
				// 	Level: slog.LevelInfo,
				// }),
				lokiHandler,
			),
		).With("app", "iceperf")

		// stop loki client and purge buffers
		defer lokiHandler.Close()

		// opts := lokirus.NewLokiHookOptions().
		// 	// Grafana doesn't have a "panic" level, but it does have a "critical" level
		// 	// https://grafana.com/docs/grafana/latest/explore/logs-integration/
		// 	WithLevelMap(lokirus.LevelMap{log.PanicLevel: "critical"}).
		// 	WithFormatter(&logrus.JSONFormatter{}).
		// 	WithStaticLabels(lokirus.Labels{
		// 		"app": "iceperftest",
		// 	})

		// if config.Logging.Loki.UseBasicAuth {
		// 	opts.WithBasicAuth(config.Logging.Loki.Username, config.Logging.Loki.Password)
		// }

		// if config.Logging.Loki.UseHeadersAuth {
		// 	httpClient := &http.Client{Transport: &transport{underlyingTransport: http.DefaultTransport, authHeaders: config.Logging.Loki.AuthHeaders}}

		// 	opts.WithHttpClient(httpClient)
		// }

		// hook := lokirus.NewLokiHookWithOpts(
		// 	config.Logging.Loki.URL,
		// 	opts,
		// 	log.InfoLevel,
		// 	log.WarnLevel,
		// 	log.ErrorLevel,
		// 	log.FatalLevel)

		// logg.AddHook(hook)

		// lokiHookConfig := &lokihook.Config{
		// 	// the loki api url
		// 	URL: config.Logging.Loki.URL,
		// 	// (optional, default: severity) the label's key to distinguish log's level, it will be added to Labels map
		// 	LevelName: "level",
		// 	// the labels which will be sent to loki, contains the {levelname: level}
		// 	Labels: map[string]string{
		// 		"app": "iceperftest",
		// 	},
		// }
		// hook, err := lokihook.NewHook(lokiHookConfig)
		// if err != nil {
		// 	log.Error(err)
		// } else {
		// 	log.AddHook(hook)
		// }

		// hook := loki.NewHook(config.Logging.Loki.URL, loki.WithLabel("app", "iceperftest"), loki.WithFormatter(&logrus.JSONFormatter{}), loki.WithLevel(log.InfoLevel))
		// defer hook.Close()

		// log.AddHook(hook)
	} else {
		handlerOpts := &slog.HandlerOptions{
			Level: loggingLevel,
		}
		logg = slog.New(slog.NewTextHandler(os.Stderr, handlerOpts))
	}
	slog.SetDefault(logg)

	if config.Timer.Enabled {
		ticker := time.NewTicker(time.Duration(config.Timer.Interval) * time.Minute)
		runTest(logg, config)
		for {
			<-ticker.C
			runTest(logg, config)
		}
	} else {
		runTest(logg, config)
	}

	return nil
}

func runTest(logg *slog.Logger, config *config.Config) error {
	// logg.SetFormatter(&log.JSONFormatter{PrettyPrint: true})

	testRunId := xid.New()
	testRunStartedAt := time.Now()

	logger := logg.With("testRunId", testRunId)

	// TODO we will make a new client for each ICE Server URL from each provider
	// get ICE servers and loop them
	ICEServers, node, err := client.GetIceServers(config, logger, testRunId)
	if err != nil {
		logger.Error("Error getting ICE servers", "err", err)
		//this should be a fatal
	}

	if node != "" {
		config.NodeID = node
	}

	config.Registry = prometheus.NewRegistry()
	// pusher := push.New(config.Logging.Loki.URL, "grafanacloud-nimbleape-prom").Gatherer(config.Registry)
	// pusher := push.New()
	// pusher.Gatherer(config.Registry)
	// promClient := promwrite.NewClient(config.Logging.Prometheus.URL)

	// TEST writing to qryn
	// if _, err := promClient.Write(
	// 	ctx.Context,
	// 	&promwrite.WriteRequest{
	// 		TimeSeries: []promwrite.TimeSeries{
	// 			{
	// 				Labels: []promwrite.Label{
	// 					{
	// 						Name:  "__name__",
	// 						Value: "test_metric",
	// 					},
	// 				},
	// 				Sample: promwrite.Sample{
	// 					Time:  time.Now(),
	// 					Value: 123,
	// 				},
	// 			},
	// 		},
	// 	},
	// 	promwrite.WriteHeaders(config.Logging.Prometheus.AuthHeaders),
	// ); err != nil {
	// 	logger.Error("Error writing to Qryn", err)
	// }
	// end TEST

	var results []*stats.Stats

	for provider, iss := range ICEServers {
		providerLogger := logger.With("Provider", provider)

		providerLogger.Info("Provider Starting")

		for _, is := range iss.IceServers {

			providerLogger.Info("URL is", "url", is)

			iceServerInfo, err := stun.ParseURI(is.URLs[0])

			if err != nil {
				providerLogger.Error("Error parsing ICE Server URL", "err", err)
				continue
			}

			runId := xid.New()

			iceServerLogger := providerLogger.With("iceServerTestRunId", runId,
				"schemeAndProtocol", iceServerInfo.Scheme.String()+"-"+iceServerInfo.Proto.String(),
			)

			iceServerLogger.Info("Starting New Client", "iceServerHost", iceServerInfo.Host,
				"iceServerProtocol", iceServerInfo.Proto.String(),
				"iceServerPort", iceServerInfo.Port,
				"iceServerScheme", iceServerInfo.Scheme.String(),
			)
			config.Logger = iceServerLogger

			config.WebRTCConfig.ICEServers = []webrtc.ICEServer{is}
			//if the ice server is a stun then set the
			testDuration := 20 * time.Second
			if iceServerInfo.Scheme == stun.SchemeTypeSTUN || iceServerInfo.Scheme == stun.SchemeTypeSTUNS {
				config.WebRTCConfig.ICETransportPolicy = webrtc.ICETransportPolicyAll
				testDuration = 2 * time.Second
			} else {
				config.WebRTCConfig.ICETransportPolicy = webrtc.ICETransportPolicyRelay
			}

			timer := time.NewTimer(testDuration)
			close := make(chan struct{})

			c, err := client.NewClient(config, iceServerInfo, provider, testRunId, testRunStartedAt, iss.DoThroughput, close)
			if err != nil {
				return err
			}

			iceServerLogger.Info("Calling Run()")
			c.Run()
			iceServerLogger.Info("Called Run(), waiting for timer", "seconds", testDuration.Seconds())
			select {
			case <-close:
				timer.Stop()
			case <-timer.C:
			}
			iceServerLogger.Info("Calling Stop()")
			c.Stop()
			<-time.After(1 * time.Second)
			iceServerLogger.Info("Finished")
			results = append(results, c.Stats)
		}
		providerLogger.Info("Provider Finished")
	}

	logger.Info("Finished Test Run")

	// c, err := client.NewClient(config)
	// if err != nil {
	// 	return nil
	// }
	// defer c.Stop()

	// c.Run()

	// util.Check(pusher.Push(config.Logging.Prometheus.URL))

	// write all metrics to qryn at once
	// mf, err := config.Registry.Gather()
	// if err != nil {
	// 	logger.Error("Error gathering metrics from registry", err)
	// }

	// if len(mf) > 0 {

	// 	timenow := time.Now()

	// 	ts := []promwrite.TimeSeries{}
	// 	for _, m := range mf {

	// 		//loop thorugh each metric
	// 		for _, met := range m.GetMetric() {
	// 			var v float64
	// 			switch m.GetType().String() {
	// 			case "GAUGE":
	// 				v = *met.Gauge.Value
	// 				//add more
	// 			}

	// 			labels := []promwrite.Label{
	// 				{
	// 					Name:  "__name__",
	// 					Value: m.GetName(),
	// 				},
	// 				{
	// 					Name:  "description",
	// 					Value: m.GetHelp(),
	// 				},
	// 				{
	// 					Name:  "type",
	// 					Value: m.GetType().String(),
	// 				},
	// 			}

	// 			for _, lp := range met.GetLabel() {
	// 				labels = append(labels, promwrite.Label{Name: *lp.Name, Value: *lp.Value})
	// 			}

	// 			ts = append(ts, promwrite.TimeSeries{
	// 				Labels: labels,
	// 				Sample: promwrite.Sample{
	// 					Time:  timenow,
	// 					Value: v,
	// 				},
	// 			})

	// 			logger.Info("got metrics", "labels", met.Label, "name", m.GetName(), "type", m.GetType(), "value", v, "unit", m.GetUnit(), "description", m.GetHelp())
	// 		}
	// 	}
	// 	_, err := promClient.Write(
	// 		ctx.Context,
	// 		&promwrite.WriteRequest{
	// 			TimeSeries: ts,
	// 		},
	// 		promwrite.WriteHeaders(config.Logging.Prometheus.AuthHeaders),
	// 	)
	// 	if err != nil {
	// 		logger.Error("Error writing to Qryn", err)
	// 	}
	// 	logger.Info("Wrote stats to prom")

	// }
	// if !config.Logging.Loki.Enabled && !config.Logging.API.Enabled {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Provider", "Scheme", "Time to candidate", "Max Throughput", "TURN Transfer Latency")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, st := range results {
		tbl.AddRow(st.Provider, st.Scheme, st.OffererTimeToReceiveCandidate, st.ThroughputMax, st.LatencyFirstPacket)
	}

	tbl.Print()
	//}
	return nil
}

func getConfig(c *cli.Context) (*config.Config, error) {
	configBody := ""
	configFile := c.String("config")
	if configFile != "" {
		content, err := os.ReadFile(configFile)
		if err != nil {
			return nil, err
		}
		configBody = string(content)
	}

	conf, err := config.NewConfig(configBody)
	if err != nil {
		return nil, err
	}

	//if we got passed in the api host and the api key then overwrite the config
	//same for timer mode
	if c.String("api-uri") != "" {
		conf.Api.Enabled = true
		conf.Api.URI = c.String("api-uri")
	}

	if c.String("api-key") != "" {
		conf.Api.Enabled = true
		conf.Api.ApiKey = c.String("api-key")
	}

	if conf.Api.Enabled && conf.Api.URI == "" {
		conf.Api.URI = "https://api.iceperf.com/api/settings"
	}

	if c.Bool("timer") {
		conf.Timer.Enabled = true
		conf.Timer.Interval = 60
	}

	if conf.Api.Enabled && conf.Api.ApiKey != "" && conf.Api.URI != "" {
		conf.UpdateConfigFromApi()
	}

	return conf, nil
}

// func setLogLevel(logger *log.Logger, level string) {
// 	switch level {
// 	case "debug":
// 		logger.SetLevel(slog.DebugLevel)
// 	case "error":
// 		logger.SetLevel(slog.ErrorLevel)
// 	case "fatal":
// 		logger.SetLevel(slog.FatalLevel)
// 	case "panic":
// 		logger.SetLevel(slog.PanicLevel)
// 	case "trace":
// 		logger.SetLevel(slog.TraceLevel)
// 	case "warn":
// 		logger.SetLevel(slog.WarnLevel)
// 	default:
// 		logger.SetLevel(slog.InfoLevel)
// 	}
// }
