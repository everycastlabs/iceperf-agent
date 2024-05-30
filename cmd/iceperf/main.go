// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/nimbleape/iceperf-agent/client"
	"github.com/nimbleape/iceperf-agent/config"
	"github.com/nimbleape/iceperf-agent/version"
	"github.com/pion/stun/v2"
	"github.com/pion/webrtc/v4"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/yukitsune/lokirus"
)

type transport struct {
	authHeaders         map[string]string
	underlyingTransport http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	for headerName, headerValue := range t.authHeaders {
		req.Header.Add(headerName, headerValue)
	}

	return t.underlyingTransport.RoundTrip(req)
}

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

	testRunId := xid.New()

	// Configure the logger
	logg := log.New()

	if config.Logging.Loki.Enabled {

		opts := lokirus.NewLokiHookOptions().
			// Grafana doesn't have a "panic" level, but it does have a "critical" level
			// https://grafana.com/docs/grafana/latest/explore/logs-integration/
			WithLevelMap(lokirus.LevelMap{log.PanicLevel: "critical"}).
			WithFormatter(&logrus.JSONFormatter{}).
			WithStaticLabels(lokirus.Labels{
				"app": "iceperf",
			})

		if config.Logging.Loki.UseBasicAuth {
			opts.WithBasicAuth(config.Logging.Loki.Username, config.Logging.Loki.Password)
		}

		if config.Logging.Loki.UseHeadersAuth {
			httpClient := &http.Client{Transport: &transport{underlyingTransport: http.DefaultTransport, authHeaders: config.Logging.Loki.AuthHeaders}}

			opts.WithHttpClient(httpClient)
		}

		hook := lokirus.NewLokiHookWithOpts(
			config.Logging.Loki.URL,
			opts,
			log.InfoLevel,
			log.WarnLevel,
			log.ErrorLevel,
			log.FatalLevel)

		logg.AddHook(hook)
	}

	// logg.SetFormatter(&log.JSONFormatter{PrettyPrint: true})
	setLogLevel(logg, config.Logging.Level)

	logger := logg.WithFields(log.Fields{
		"testRunId": testRunId,
	})

	// TODO we will make a new client for each ICE Server URL from each provider
	// get ICE servers and loop them
	ICEServers, err := client.GetIceServers(config)
	if err != nil {
		logger.Fatal("Error getting ICE servers")
	}
	// log.WithFields(log.Fields{
	// 	"ICEServers": ICEServers,
	// }).Info("ICE Servers in use")

	for provider, iss := range ICEServers {
		providerLogger := logger.WithFields(log.Fields{
			"Provider": provider,
		})

		providerLogger.Info("Provider Starting")

		for _, is := range iss {

			iceServerInfo, err := stun.ParseURI(is.URLs[0])

			if err != nil {
				return err
			}

			runId := xid.New()

			iceServerLogger := providerLogger.WithFields(log.Fields{
				"iceServerTestRunId": runId,
				"schemeAndProtocol":  iceServerInfo.Scheme.String() + "-" + iceServerInfo.Proto.String(),
			})

			iceServerLogger.WithFields(log.Fields{
				"iceServerHost":     iceServerInfo.Host,
				"iceServerProtocol": iceServerInfo.Proto.String(),
				"iceServerPort":     iceServerInfo.Port,
				"iceServerScheme":   iceServerInfo.Scheme.String(),
			}).Info("Starting New Client")
			config.Logger = iceServerLogger

			config.WebRTCConfig.ICEServers = []webrtc.ICEServer{is}
			//if the ice server is a stun then set the
			if iceServerInfo.Scheme == stun.SchemeTypeSTUN || iceServerInfo.Scheme == stun.SchemeTypeSTUNS {
				config.WebRTCConfig.ICETransportPolicy = webrtc.ICETransportPolicyAll
			} else {
				config.WebRTCConfig.ICETransportPolicy = webrtc.ICETransportPolicyRelay
			}

			timer := time.NewTimer(20 * time.Second)
			c, err := client.NewClient(config, iceServerInfo)
			if err != nil {
				return err
			}

			iceServerLogger.Info("Calling Run()")
			c.Run()
			iceServerLogger.Info("Called Run(), waiting for timer 10 seconds")
			<-timer.C
			iceServerLogger.Info("Calling Stop()")
			c.Stop()
			<-time.After(2 * time.Second)
			iceServerLogger.Info("Finished")
		}
		providerLogger.Info("Provider Finished")
	}

	// c, err := client.NewClient(config)
	// if err != nil {
	// 	return nil
	// }
	// defer c.Stop()

	// c.Run()
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

	return conf, nil
}

func setLogLevel(logger *log.Logger, level string) {
	switch level {
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	case "fatal":
		logger.SetLevel(log.FatalLevel)
	case "panic":
		logger.SetLevel(log.PanicLevel)
	case "trace":
		logger.SetLevel(log.TraceLevel)
	case "warn":
		logger.SetLevel(log.WarnLevel)
	default:
		logger.SetLevel(log.InfoLevel)
	}
}
