package main

import (
	"flag"
	"fmt"
	"os"
	"net/http"
	"time"

	"example.org/config"
	"example.org/curl"
	"example.org/dns"
	"example.org/logger"
	"example.org/ping"
	"example.org/tcping"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const version string = "0.0.1"

var (
	showVersion        = flag.Bool("version", false, "Print version information.")
	listenAddress      = flag.String("web.listen-address", "0.0.0.0:9909", "Address on which to expose metrics and web interface.")
	metricsPath        = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	level              = flag.String("level", "info", "Set logging verbose level")
	timeout            = flag.Int("Timeout", 5, "Set default test timeout value")
	configFile         = flag.String("config.file", "config.yaml", "Path to config file (required)")
)

var (
	registry = prometheus.NewRegistry()
	testResults = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sla_result",
			Help: "Total number of test results by type, and target.",
		},
		[]string{"test", "target"},
	)
	testDurations = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "sla_duration",
            Help: "Duration of tests by test and target.",
        },
        []string{"test", "target"},
    )
)

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: sla_exporter [ ... ]\n\nParameters:")
		fmt.Println()
		flag.PrintDefaults()
	}
}


func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	err := initialize()
	if err != nil {
		logger.Log.Fatal("could not initialize exporter.",
			zap.Error(err),
		)
	}

	// Register metrics with the global Prometheus registry
    registry.MustRegister(testResults)
    registry.MustRegister(testDurations)
	dns.Register(registry)
	curl.Register(registry)
	ping.Register(registry)
	tcping.Register(registry)

	startServer()
}


func initialize() error {
	var err error
	c := config.New()
	c.ConfigFile = *configFile

	c, err = loadConfig(c)
	if err != nil {
		return err
	}

	config.AppConfig = c
	logger.Config()

	return nil
}


func printVersion() {
	fmt.Println("sla_exporter")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("Author(s): slashdoom (Patrick Ryon)")
	fmt.Println("Metric exporter for simple network SLA tests")
}


func startServer() {
	logger.Log.Info("starting sla_exporter",
		zap.String("version", version),
	)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
<html>
  <head>
    <title>SLA Exporter (Version ` + version + `)</title>
  </head>
  <body>
    <h1>SLA Exporter</h1>
    <p><a href="` + *metricsPath + `">Metrics</a></p>
    <h2>More information:</h2>
    <p><a href="https://github.com/slashdoom/sla_exporter">github.com/slashdoom/sla_exporter</a></p>
  </body>
</html>
`))
	})
	http.HandleFunc(*metricsPath, handleMetricsRequest)

	logger.Log.Info("listening for connections...",
		zap.String("metricsPath", *metricsPath),
		zap.String("listenAddress", *listenAddress),
	)
	http.ListenAndServe(*listenAddress, nil)
}


func handleMetricsRequest(w http.ResponseWriter, r *http.Request) {
	reg := registry

	startTimeTotal := time.Now()

	// Run cURL tests
    startTimeCurl := time.Now()
	for _, curlTest := range config.AppConfig.CurlTests {
		result, _ := curl.Test(curlTest)
		logger.Log.Debug("cURL result",
            zap.String("result", fmt.Sprintf("%v", result.Completed)),
            zap.String("duration", fmt.Sprintf("%s", result.Duration)),
        )
        curl.Record(fmt.Sprintf("%s", curlTest.URL), "GET", result)
	}
    recordTestResult("curl", "all_curl_tests", true, time.Since(startTimeCurl))

	// Run DNS tests
    startTimeDns := time.Now()
	for _, dnsTest := range config.AppConfig.DnsTests {
		result, _ := dns.Test(dnsTest)
		logger.Log.Debug("DNS result",
            zap.String("result", fmt.Sprintf("%v", result.Completed)),
            zap.String("duration", fmt.Sprintf("%s", result.Duration)),
        )
		serverName := ""
		if dnsTest.Server != "" {
			serverName = fmt.Sprintf("%s", dnsTest.Server)
		}
        dns.Record(fmt.Sprintf("%s", dnsTest.Host), fmt.Sprintf("%s", serverName), result)
	}
    recordTestResult("dns", "all_dns_tests", true, time.Since(startTimeDns))

	// Run ping tests
    startTimePing := time.Now()
	for _, pingTest := range config.AppConfig.PingTests {
		result, _ := ping.Test(pingTest)
        logger.Log.Debug("Ping result",
            zap.String("result", fmt.Sprintf("%v", result.Completed)),
            zap.String("duration", fmt.Sprintf("%s", result.Duration)),
        )
		ping.Record(fmt.Sprintf("%s", pingTest.Host), result)
	}
    recordTestResult("ping", "all_ping_tests", true, time.Since(startTimePing))

	// Run tcping tests
    startTimeTcping := time.Now()
	for _, tcpingTest := range config.AppConfig.TcpingTests {
		result, _ := tcping.Test(tcpingTest)
        logger.Log.Debug("TCPing result",
            zap.String("result", fmt.Sprintf("%v", result.Completed)),
            zap.String("duration", fmt.Sprintf("%v", result.Duration)),
        )
        tcping.Record(fmt.Sprintf("%s", tcpingTest.Host), fmt.Sprintf("%d", tcpingTest.Port), result)
	}
    recordTestResult("tcping", "all_tcping_tests", true, time.Since(startTimeTcping))

	recordTestResult("all_tests", "total", true, time.Since(startTimeTotal))

	promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(w, r)
}


func recordTestResult(test string, target string, result bool, duration time.Duration) {
	succeeded := 0.00
	if result {
		succeeded = 1.00
	}
	testResults.WithLabelValues(test, target).Set(succeeded)
	testDurations.WithLabelValues(test, target).Set(duration.Seconds())
}