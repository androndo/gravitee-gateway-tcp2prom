package main

import (
	"fmt"
	"github.com/firstrow/tcp_server"
	jsoniter "github.com/json-iterator/go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"time"
)

var (
	labels = []string{"host", "api", "application", "statusCodeFamily", "httpMethod"}
)

var (
	sessionCounter = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gateway_exporter_tcp_session_count",
		Help: "Current connected tcp sessions",
	})
	// todo: maybe change to vec and split by types of events
	exporterRequestFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_exporter_tcp_requests_failed",
		Help: "How many requests couldn't unmarshall from json",
	})
	exporterRequestProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_exporter_tcp_requests_processed",
		Help: "How many requests processed successfully",
	})
	exporterRequestTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_exporter_tcp_requests_total",
		Help: "How many requests received",
	})
	exporterElapsedTime = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "gateway_exporter_processing_elapsed_time_seconds",
		Help:       "Time from input message to output metrics.",
		Objectives: percentilesMap,
	})
	gatewayResponseCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gateway_http_responses_count",
		Help: "How many HTTP requests processed, partitioned by status host, uri, status code family and http method.\",",
	}, labels)
	percentilesMap  = map[float64]float64{0.5: 0.01, 0.75: 0.01, 0.95: 0.01, 0.99: 0.01}
	apiResponseTime = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "gateway_api_response_time",
		Help:       "How long answered API.\",",
		Objectives: percentilesMap,
	}, labels)
	gatewayLatencyTime = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "gateway_proxy_latency_time",
		Help:       "How long spent gateway proxy.\",",
		Objectives: percentilesMap,
	}, labels)
	gatewayRequestSize = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "gateway_http_requests_size",
		Help:       "Content length in http requests.\",",
		Objectives: percentilesMap,
	}, labels)
	gatewayResponseSize = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "gateway_http_responses_size",
		Help:       "Content length in http responses.\",",
		Objectives: percentilesMap,
	}, labels)
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	// exporter
	prometheus.MustRegister(sessionCounter)
	prometheus.MustRegister(exporterRequestFailed)
	prometheus.MustRegister(exporterRequestProcessed)
	prometheus.MustRegister(exporterRequestTotal)
	prometheus.MustRegister(exporterElapsedTime)
	// http proxy
	prometheus.MustRegister(gatewayResponseCount)
	prometheus.MustRegister(apiResponseTime)
	prometheus.MustRegister(gatewayLatencyTime)
	prometheus.MustRegister(gatewayRequestSize)
	prometheus.MustRegister(gatewayResponseSize)
}

func onNewClient(c *tcp_server.Client) {
	sessionCounter.Inc()
	log.Info().Stringer("from", c.Conn().RemoteAddr()).Msg("New connection")
}

func onNewMessage(c *tcp_server.Client, message string) {
	defer observeExporterElapsedTime(time.Now())

	log.Debug().
		Stringer("from", c.Conn().RemoteAddr()).
		Str("raw_message", message).
		Msg("New message")

	var out mappingStruct

	if err := json.UnmarshalFromString(message, &out); err != nil {
		exporterRequestFailed.Inc()
		log.Error().Err(err).Msg("Error unmarshal")
		return
	}

	if out.HttpMethod == "" {
		exporterRequestProcessed.Inc()
		log.Info().Msg("Other type, will be skipped")
		return
	}

	var h HttpEvent

	if err := json.UnmarshalFromString(message, &h); err != nil {
		exporterRequestFailed.Inc()
		log.Error().Err(err).Msg("Error unmarshal http")
		return
	}

	exporterRequestProcessed.Inc()
	labels := []string{h.Host, h.Api, h.Application, fmt.Sprintf("%dXX", h.Status/100), h.HttpMethod}
	gatewayResponseCount.WithLabelValues(labels...).Inc()
	gatewayRequestSize.WithLabelValues(labels...).Observe(float64(h.RequestContentLength))
	gatewayResponseSize.WithLabelValues(labels...).Observe(float64(h.ResponseContentLength))
	gatewayLatencyTime.WithLabelValues(labels...).Observe(float64(h.ProxyLatencyMs) / 1000)
	apiResponseTime.WithLabelValues(labels...).Observe(float64(h.ApiResponseTimeMs) / 1000)
	log.Debug().Interface("http_event", h).Msg("Metrics updated from http event")
	return
}

func onClientConnectionClosed(c *tcp_server.Client, _ error) {
	sessionCounter.Dec()
	log.Info().Stringer("from", c.Conn().RemoteAddr()).Msg("Connection closed from")
}

func observeExporterElapsedTime(start time.Time) {
	elapsed := time.Now().Sub(start).Seconds()
	exporterElapsedTime.Observe(elapsed)
	exporterRequestTotal.Inc()
}

func getEnvOrDefault(key, _default string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	} else {
		return _default
	}
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if l, e := zerolog.ParseLevel(getEnvOrDefault("LOG_LEVEL", "info")); e == nil {
		zerolog.SetGlobalLevel(l)
		log.Info().Msg("Log level is " + l.String())
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Info().Msg("Unknown LOG_LEVEL, used default info level")
	}

	tcpAddress := getEnvOrDefault("TCP_ADDR", ":8123")
	log.Info().Msg("tcp address is " + tcpAddress)
	server := tcp_server.New(tcpAddress)
	server.OnNewClient(onNewClient)
	server.OnNewMessage(onNewMessage)
	server.OnClientConnectionClosed(onClientConnectionClosed)

	go func() {
		path := getEnvOrDefault("METRICS_PATH", "/metrics")
		http.Handle(path, promhttp.Handler())
		// ListenAndServe always returns non-nil error
		address := getEnvOrDefault("METRICS_ADDR", ":8080")
		log.Info().Msg("metrics address is " + address + " and path is " + path)
		log.Fatal().Err(http.ListenAndServe(address, nil))
	}()

	server.Listen()
	// todo: handle os signals
}
